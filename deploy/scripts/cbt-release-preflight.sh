#!/usr/bin/env bash
set -euo pipefail

DEFAULT_FLUTTER_BIN_DIR="/home/servermtsn2kolut/development/flutter/bin"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
DEFAULT_OUTPUT_DIR="${REPO_ROOT}/tmp/cbt-release-preflight-$(date +%Y%m%d-%H%M%S)"
EVIDENCE_TEMPLATE="docs/cbt-release-evidence-template.md"

OUTPUT_DIR=""
RUN_WEB_DOCS_GUARD=0
RUN_WEB_CHECK=0
RUN_FLUTTER_DOCTOR=0
RUN_FLUTTER_ANALYZE=0
RUN_FLUTTER_TEST=0
RUN_GO_TEST=0
RUN_GO_BUILD=0
TEST_ONLY_WRITE_SECRET_LEAK=0

WEB_DOCS_GUARD_DISPLAY="npm run test:unit -- src/lib/cbt/proposal-integration-docs.test.ts"
WEB_CHECK_DISPLAY="npm run check"
GO_TEST_DISPLAY="go test ./..."
GO_BUILD_DISPLAY="go build -o /dev/null ./cmd/api"
FLUTTER_DOCTOR_DISPLAY="flutter doctor"
FLUTTER_ANALYZE_DISPLAY="flutter analyze"
FLUTTER_TEST_DISPLAY="flutter test"

CHECK_NAMES=()
CHECK_STATUSES=()
CHECK_DETAILS=()
CHECK_LOGS=()
OVERALL_STATUS="pass"
SECRET_SCAN_STATUS="pending"
SECRET_SCAN_FILES=()
SECRET_SCAN_FINDINGS=()

usage() {
	cat <<'USAGE'
Usage: deploy/scripts/cbt-release-preflight.sh [options]

Collect read-only CBT release evidence into tmp/ by default.

Options:
  --output <dir>             Write markdown/json/log evidence to this directory.
  --run-web-docs-guard      Run the targeted CBT docs Vitest guard.
  --run-web-check           Run SvelteKit/svelte-check for web-admin.
  --run-flutter-doctor      Run flutter doctor if Flutter is available.
  --run-flutter-analyze     Run flutter analyze if Flutter is available.
  --run-flutter-test        Run flutter test if Flutter is available.
  --run-go-test             Run Go unit tests for services/core-api.
  --run-go-build            Run a core-api build check to /dev/null.
  --all-safe-checks         Enable every optional non-destructive check above.
  --test-only-write-secret-leak
                            Test-only hook: write synthetic secret-like evidence
                            under the output directory so the scanner must fail.
  -h, --help                Show this help.

This script writes evidence only. It does not perform deployment, process
manager changes, database migrations, SQL mutation, or runtime state changes.
USAGE
}

fail_usage() {
	printf 'ERROR: %s\n\n' "$1" >&2
	usage >&2
	exit 2
}

while [ "$#" -gt 0 ]; do
	case "$1" in
		--output)
			[ "$#" -ge 2 ] || fail_usage "--output requires a directory"
			OUTPUT_DIR="$2"
			shift 2
			;;
		--run-web-docs-guard)
			RUN_WEB_DOCS_GUARD=1
			shift
			;;
		--run-web-check)
			RUN_WEB_CHECK=1
			shift
			;;
		--run-flutter-doctor)
			RUN_FLUTTER_DOCTOR=1
			shift
			;;
		--run-flutter-analyze)
			RUN_FLUTTER_ANALYZE=1
			shift
			;;
		--run-flutter-test)
			RUN_FLUTTER_TEST=1
			shift
			;;
		--run-go-test)
			RUN_GO_TEST=1
			shift
			;;
		--run-go-build)
			RUN_GO_BUILD=1
			shift
			;;
		--all-safe-checks)
			RUN_WEB_DOCS_GUARD=1
			RUN_WEB_CHECK=1
			RUN_FLUTTER_DOCTOR=1
			RUN_FLUTTER_ANALYZE=1
			RUN_FLUTTER_TEST=1
			RUN_GO_TEST=1
			RUN_GO_BUILD=1
			shift
			;;
		--test-only-write-secret-leak)
			TEST_ONLY_WRITE_SECRET_LEAK=1
			shift
			;;
		-h|--help)
			usage
			exit 0
			;;
		*)
			fail_usage "unknown option: $1"
			;;
	esac
done

resolve_output_dir() {
	local requested="$1"
	if [ -z "$requested" ]; then
		printf '%s\n' "$DEFAULT_OUTPUT_DIR"
	elif [[ "$requested" == /* ]]; then
		printf '%s\n' "$requested"
	else
		printf '%s/%s\n' "$REPO_ROOT" "$requested"
	fi
}

OUTPUT_DIR="$(resolve_output_dir "$OUTPUT_DIR")"
LOG_DIR="${OUTPUT_DIR}/logs"
MARKDOWN_REPORT="${OUTPUT_DIR}/cbt-release-preflight.md"
JSON_REPORT="${OUTPUT_DIR}/cbt-release-preflight.json"
MANIFEST_REPORT="${OUTPUT_DIR}/cbt-release-manifest.json"

if [ -e "$OUTPUT_DIR" ] && [ ! -d "$OUTPUT_DIR" ]; then
	printf 'ERROR: output path exists but is not a directory: %s\n' "$OUTPUT_DIR" >&2
	exit 2
fi

if [ -d "$OUTPUT_DIR" ] && [ -n "$(find "$OUTPUT_DIR" -mindepth 1 -maxdepth 1 -print -quit)" ]; then
	printf 'ERROR: output directory must be empty: %s\n' "$OUTPUT_DIR" >&2
	exit 2
fi

mkdir -p "$LOG_DIR"

redact_stream() {
	sed -E \
		-e 's/(Authorization:[[:space:]]*[Bb]earer[[:space:]]+)[^[:space:]]+/\1[REDACTED]/g' \
		-e 's/((token|password|secret|api[ _-]?key|answer[ _-]?key)[[:space:]]*[:=][[:space:]]*)[^[:space:],}]+/\1[REDACTED]/Ig'
}

json_escape() {
	local value="$1"
	value="${value//\\/\\\\}"
	value="${value//\"/\\\"}"
	value="${value//$'\n'/\\n}"
	value="${value//$'\r'/\\r}"
	value="${value//$'\t'/\\t}"
	printf '%s' "$value"
}

relative_output_path() {
	local path="$1"
	if [ -z "$path" ]; then
		printf ''
	elif [[ "$path" == "$OUTPUT_DIR/"* ]]; then
		printf '%s\n' "${path#"$OUTPUT_DIR/"}"
	else
		printf '%s\n' "$path"
	fi
}

record_check() {
	local name="$1"
	local status="$2"
	local detail="$3"
	local log_path="${4:-}"

	CHECK_NAMES+=("$name")
	CHECK_STATUSES+=("$status")
	CHECK_DETAILS+=("$detail")
	CHECK_LOGS+=("$log_path")

	if [ "$status" = "fail" ]; then
		OVERALL_STATUS="failed"
	fi
}

relative_log_path() {
	relative_output_path "$1"
}

scan_secret_pattern() {
	local file="$1"
	local rel_path="$2"
	local label="$3"
	local pattern="$4"
	local line_no=""

	while IFS=: read -r line_no _; do
		[ -n "$line_no" ] || continue
		SECRET_SCAN_FINDINGS+=("${rel_path}:${line_no}:${label}")
	done < <(LC_ALL=C grep -EIn "$pattern" "$file" 2>/dev/null | grep -Fv '[REDACTED]' || true)
}

scan_generated_evidence_for_secrets() {
	SECRET_SCAN_FILES=()
	SECRET_SCAN_FINDINGS=()

	local file=""
	local rel_path=""
	local auth_pattern='Authorization:[[:space:]]*[Bb]earer[[:space:]]+[^[:space:]]+'
	local key_pattern="[\"']?(token|password|secret|api[ _-]?key|answer[ _-]?key)[\"']?[[:space:]]*[:=][[:space:]]*[\"']?[^\"'[:space:],}]+"

	while IFS= read -r file; do
		[ -n "$file" ] || continue
		rel_path="$(relative_output_path "$file")"
		SECRET_SCAN_FILES+=("$rel_path")
		scan_secret_pattern "$file" "$rel_path" "authorization-bearer" "$auth_pattern"
		scan_secret_pattern "$file" "$rel_path" "sensitive-assignment" "$key_pattern"
	done < <(find "$OUTPUT_DIR" -type f \( -name '*.md' -o -name '*.json' -o -name '*.log' \) ! -path "$MANIFEST_REPORT" -print | LC_ALL=C sort)

	if [ "${#SECRET_SCAN_FINDINGS[@]}" -gt 0 ]; then
		SECRET_SCAN_STATUS="fail"
		printf 'ERROR: secret scan failed; generated evidence contains unredacted sensitive-looking values:\n' >&2
		local finding=""
		for finding in "${SECRET_SCAN_FINDINGS[@]}"; do
			printf '  %s\n' "$finding" >&2
		done
		return 1
	fi

	SECRET_SCAN_STATUS="pass"
	return 0
}

sha256_file() {
	local file="$1"
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$file" | awk '{print $1}'
	elif command -v shasum >/dev/null 2>&1; then
		shasum -a 256 "$file" | awk '{print $1}'
	else
		printf 'ERROR: sha256sum or shasum is required to write %s\n' "$MANIFEST_REPORT" >&2
		exit 2
	fi
}

file_size_bytes() {
	wc -c < "$1" | tr -d '[:space:]'
}

artifact_kind() {
	case "$1" in
		*.md) printf 'markdown' ;;
		*.json) printf 'json' ;;
		*.log) printf 'log' ;;
		*) printf 'artifact' ;;
	esac
}

write_manifest() {
	local artifacts=()
	local file=""
	while IFS= read -r file; do
		[ -n "$file" ] || continue
		artifacts+=("$file")
	done < <(find "$OUTPUT_DIR" -type f \( -name '*.md' -o -name '*.json' -o -name '*.log' \) ! -path "$MANIFEST_REPORT" -print | LC_ALL=C sort)

	{
		printf '{\n'
		printf '  "generated_at": "%s",\n' "$(json_escape "$GENERATED_AT")"
		printf '  "repo_root": "%s",\n' "$(json_escape "$REPO_ROOT")"
		printf '  "branch": "%s",\n' "$(json_escape "${BRANCH_NAME:-unknown}")"
		printf '  "commit": "%s",\n' "$(json_escape "${HEAD_SHA:-unknown}")"
		printf '  "output_dir": "%s",\n' "$(json_escape "$OUTPUT_DIR")"
		printf '  "algorithm": "sha256",\n'
		printf '  "secret_scan": {\n'
		printf '    "status": "%s",\n' "$(json_escape "$SECRET_SCAN_STATUS")"
		printf '    "scanned_files": [\n'
		for i in "${!SECRET_SCAN_FILES[@]}"; do
			printf '      "%s"' "$(json_escape "${SECRET_SCAN_FILES[$i]}")"
			if [ "$i" -eq "$((${#SECRET_SCAN_FILES[@]} - 1))" ]; then
				printf '\n'
			else
				printf ',\n'
			fi
		done
		printf '    ],\n'
		printf '    "findings": [\n'
		for i in "${!SECRET_SCAN_FINDINGS[@]}"; do
			printf '      "%s"' "$(json_escape "${SECRET_SCAN_FINDINGS[$i]}")"
			if [ "$i" -eq "$((${#SECRET_SCAN_FINDINGS[@]} - 1))" ]; then
				printf '\n'
			else
				printf ',\n'
			fi
		done
		printf '    ]\n'
		printf '  },\n'
		printf '  "artifacts": [\n'
		for i in "${!artifacts[@]}"; do
			local rel_path=""
			local kind=""
			local size=""
			local checksum=""
			rel_path="$(relative_output_path "${artifacts[$i]}")"
			kind="$(artifact_kind "$rel_path")"
			size="$(file_size_bytes "${artifacts[$i]}")"
			checksum="$(sha256_file "${artifacts[$i]}")"
			printf '    {\n'
			printf '      "path": "%s",\n' "$(json_escape "$rel_path")"
			printf '      "kind": "%s",\n' "$(json_escape "$kind")"
			printf '      "bytes": %s,\n' "$size"
			printf '      "sha256": "%s"\n' "$(json_escape "$checksum")"
			if [ "$i" -eq "$((${#artifacts[@]} - 1))" ]; then
				printf '    }\n'
			else
				printf '    },\n'
			fi
		done
		printf '  ]\n'
		printf '}\n'
	} > "$MANIFEST_REPORT"
}

write_markdown_report() {
	{
		printf '# CBT Release Preflight Evidence\n\n'
		printf '%s\n' "- Generated at: \`${GENERATED_AT}\`"
		printf '%s\n' "- Repository: \`${REPO_ROOT}\`"
		printf '%s\n' "- Branch: \`${BRANCH_NAME:-unknown}\`"
		printf '%s\n' "- Commit: \`${HEAD_SHA:-unknown}\`"
		printf '%s\n' "- Overall status: \`${OVERALL_STATUS}\`"
		printf '%s\n' "- Evidence template: \`${EVIDENCE_TEMPLATE}\`"
		printf '%s\n\n' "- Output directory: \`${OUTPUT_DIR}\`"

		printf '## Safety Scope\n\n'
		printf '%s\n' '- This is read-only release evidence collection; Phase 8 manifest/checksum/secret-scan hardening only strengthens this verifier evidence.'
		printf '%s\n' '- It writes markdown/json/log evidence only under the output directory.'
		printf '%s\n' '- It also writes `cbt-release-manifest.json` with SHA-256 checksums for generated markdown/json/log evidence.'
		printf '%s\n' '- It fails if generated evidence still contains unredacted token/password/secret/API key/answer key patterns.'
		printf '%s\n' '- It does not perform deployment, process manager changes, database migrations, SQL mutation, or runtime state changes.'
		printf '%s\n\n' '- Flutter runtime evidence remains scoped to `services/core-api` `/api/exam/*`.'

		printf '## Checks\n\n'
		printf '| Check | Status | Detail | Log |\n'
		printf '|-------|--------|--------|-----|\n'
		for i in "${!CHECK_NAMES[@]}"; do
			printf '| `%s` | `%s` | %s | %s |\n' \
				"${CHECK_NAMES[$i]}" \
				"${CHECK_STATUSES[$i]}" \
				"${CHECK_DETAILS[$i]//|/\\|}" \
				"${CHECK_LOGS[$i]:-}"
		done

		printf '\n## Notes\n\n'
		printf '%s\n' "- Skipped optional checks are not release approval; record the reason in \`${EVIDENCE_TEMPLATE}\`."
		printf '%s\n' "- Verify \`$(basename "$MANIFEST_REPORT")\` before archiving the evidence bundle."
		printf '%s\n' '- If a requested check or the secret scan failed, treat this bundle as a blocker until reviewed.'
	} > "$MARKDOWN_REPORT"
}

write_json_report() {
	{
		printf '{\n'
		printf '  "generated_at": "%s",\n' "$(json_escape "$GENERATED_AT")"
		printf '  "repo_root": "%s",\n' "$(json_escape "$REPO_ROOT")"
		printf '  "branch": "%s",\n' "$(json_escape "${BRANCH_NAME:-unknown}")"
		printf '  "commit": "%s",\n' "$(json_escape "${HEAD_SHA:-unknown}")"
		printf '  "overall_status": "%s",\n' "$(json_escape "$OVERALL_STATUS")"
		printf '  "evidence_template": "%s",\n' "$(json_escape "$EVIDENCE_TEMPLATE")"
		printf '  "output_dir": "%s",\n' "$(json_escape "$OUTPUT_DIR")"
		printf '  "manifest": "%s",\n' "$(json_escape "$(basename "$MANIFEST_REPORT")")"
		printf '  "checks": [\n'
		for i in "${!CHECK_NAMES[@]}"; do
			printf '    {\n'
			printf '      "name": "%s",\n' "$(json_escape "${CHECK_NAMES[$i]}")"
			printf '      "status": "%s",\n' "$(json_escape "${CHECK_STATUSES[$i]}")"
			printf '      "detail": "%s",\n' "$(json_escape "${CHECK_DETAILS[$i]}")"
			printf '      "log": "%s"\n' "$(json_escape "${CHECK_LOGS[$i]:-}")"
			if [ "$i" -eq "$((${#CHECK_NAMES[@]} - 1))" ]; then
				printf '    }\n'
			else
				printf '    },\n'
			fi
		done
		printf '  ]\n'
		printf '}\n'
	} > "$JSON_REPORT"
}

run_logged_command() {
	local name="$1"
	local workdir="$2"
	local display="$3"
	shift 3

	local tmp_log="${LOG_DIR}/${name}.tmp"
	local log_path="${LOG_DIR}/${name}.log"
	local status=0

	set +e
	(
		cd "$workdir"
		GIT_OPTIONAL_LOCKS=0 "$@"
	) > "$tmp_log" 2>&1
	status=$?
	set -e

	redact_stream < "$tmp_log" > "$log_path"
	rm -f "$tmp_log"

	if [ "$status" -eq 0 ]; then
		record_check "$name" "pass" "$display" "$(relative_log_path "$log_path")"
	else
		record_check "$name" "fail" "$display exited $status" "$(relative_log_path "$log_path")"
	fi
}

record_tool_availability() {
	local tool="$1"
	local name="tool-${tool}"
	local found=""

	if found="$(command -v "$tool" 2>/dev/null)"; then
		record_check "$name" "pass" "$found" ""
	else
		record_check "$name" "skipped" "not found on PATH" ""
	fi
}

detect_flutter() {
	if [ -x "${DEFAULT_FLUTTER_BIN_DIR}/flutter" ]; then
		printf '%s\n' "${DEFAULT_FLUTTER_BIN_DIR}/flutter"
	elif command -v flutter >/dev/null 2>&1; then
		command -v flutter
	else
		printf ''
	fi
}

run_optional_command() {
	local enabled="$1"
	local name="$2"
	local workdir="$3"
	local display="$4"
	local required_tool="$5"
	shift 5

	if [ "$enabled" -ne 1 ]; then
		record_check "$name" "skipped" "not requested; enable with the matching flag" ""
		return
	fi

	if [ -n "$required_tool" ]; then
		if [[ "$required_tool" == */* ]]; then
			if [ ! -x "$required_tool" ]; then
				record_check "$name" "fail" "requested but tool is not executable: $required_tool" ""
				return
			fi
		elif ! command -v "$required_tool" >/dev/null 2>&1; then
			record_check "$name" "fail" "requested but tool is not available: $required_tool" ""
			return
		fi
	fi

	run_logged_command "$name" "$workdir" "$display" "$@"
}

GENERATED_AT="$(date -Iseconds)"
HEAD_SHA="$(GIT_OPTIONAL_LOCKS=0 git -C "$REPO_ROOT" rev-parse HEAD 2>/dev/null || true)"
BRANCH_NAME="$(GIT_OPTIONAL_LOCKS=0 git -C "$REPO_ROOT" branch --show-current 2>/dev/null || true)"
FLUTTER_CMD="$(detect_flutter)"

record_tool_availability "git"
record_tool_availability "node"
record_tool_availability "npm"
record_tool_availability "go"

if [ -n "$FLUTTER_CMD" ]; then
	record_check "tool-flutter" "pass" "$FLUTTER_CMD" ""
else
	record_check "tool-flutter" "skipped" "not found at ${DEFAULT_FLUTTER_BIN_DIR}/flutter or on PATH" ""
fi

run_logged_command "git-head" "$REPO_ROOT" "git log -1 --format=%H%n%an%n%aI%n%s" git log -1 "--format=%H%n%an%n%aI%n%s"
run_logged_command "git-status" "$REPO_ROOT" "git status --short --branch" git status --short --branch
run_logged_command "git-diff-check" "$REPO_ROOT" "git diff --check" git diff --check

run_optional_command "$RUN_WEB_DOCS_GUARD" "web-docs-guard" "${REPO_ROOT}/apps/web-admin" "$WEB_DOCS_GUARD_DISPLAY" "npm" npm run test:unit -- src/lib/cbt/proposal-integration-docs.test.ts
run_optional_command "$RUN_WEB_CHECK" "web-check" "${REPO_ROOT}/apps/web-admin" "$WEB_CHECK_DISPLAY" "npm" npm run check
run_optional_command "$RUN_GO_TEST" "go-test" "${REPO_ROOT}/services/core-api" "$GO_TEST_DISPLAY" "go" go test ./...
run_optional_command "$RUN_GO_BUILD" "go-build" "${REPO_ROOT}/services/core-api" "$GO_BUILD_DISPLAY" "go" go build -o /dev/null ./cmd/api

if [ -n "$FLUTTER_CMD" ]; then
	run_optional_command "$RUN_FLUTTER_DOCTOR" "flutter-doctor" "${REPO_ROOT}/apps/mobile" "$FLUTTER_DOCTOR_DISPLAY" "$FLUTTER_CMD" "$FLUTTER_CMD" doctor
	run_optional_command "$RUN_FLUTTER_ANALYZE" "flutter-analyze" "${REPO_ROOT}/apps/mobile" "$FLUTTER_ANALYZE_DISPLAY" "$FLUTTER_CMD" "$FLUTTER_CMD" analyze
	run_optional_command "$RUN_FLUTTER_TEST" "flutter-test" "${REPO_ROOT}/apps/mobile" "$FLUTTER_TEST_DISPLAY" "$FLUTTER_CMD" "$FLUTTER_CMD" test
else
	if [ "$RUN_FLUTTER_DOCTOR" -eq 1 ]; then
		record_check "flutter-doctor" "fail" "requested but Flutter SDK is unavailable at ${DEFAULT_FLUTTER_BIN_DIR}/flutter or PATH" ""
	else
		record_check "flutter-doctor" "skipped" "not requested; Flutter SDK unavailable" ""
	fi
	if [ "$RUN_FLUTTER_ANALYZE" -eq 1 ]; then
		record_check "flutter-analyze" "fail" "requested but Flutter SDK is unavailable at ${DEFAULT_FLUTTER_BIN_DIR}/flutter or PATH" ""
	else
		record_check "flutter-analyze" "skipped" "not requested; Flutter SDK unavailable" ""
	fi
	if [ "$RUN_FLUTTER_TEST" -eq 1 ]; then
		record_check "flutter-test" "fail" "requested but Flutter SDK is unavailable at ${DEFAULT_FLUTTER_BIN_DIR}/flutter or PATH" ""
	else
		record_check "flutter-test" "skipped" "not requested; Flutter SDK unavailable" ""
	fi
fi

if [ "$TEST_ONLY_WRITE_SECRET_LEAK" -eq 1 ]; then
	printf 'api_key = test-only-unredacted-value\n' > "${LOG_DIR}/test-only-secret-leak.log"
fi

write_markdown_report
write_json_report

if scan_generated_evidence_for_secrets; then
	record_check "secret-scan" "pass" "generated evidence scanned for unredacted token/password/secret/API key/answer key patterns" ""
else
	record_check "secret-scan" "fail" "generated evidence contains unredacted sensitive-looking values; review stderr findings by file, line, and pattern" ""
fi

write_markdown_report
write_json_report
write_manifest

printf 'CBT release preflight evidence written:\n'
printf '  Markdown: %s\n' "$MARKDOWN_REPORT"
printf '  JSON:     %s\n' "$JSON_REPORT"
printf '  Manifest: %s\n' "$MANIFEST_REPORT"
printf '  Logs:     %s\n' "$LOG_DIR"

if [ "$OVERALL_STATUS" = "failed" ]; then
	exit 1
fi
