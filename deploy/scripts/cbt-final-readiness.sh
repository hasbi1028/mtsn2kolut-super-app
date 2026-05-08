#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
DEFAULT_OUTPUT_DIR="${REPO_ROOT}/tmp/cbt-final-readiness"

OUTPUT_DIR=""
RUN_OPS_HEALTH=0
MANUAL_DEVICE_MATRIX_COMPLETE=0
MANUAL_OPERATOR_REHEARSAL_COMPLETE=0
MANUAL_FINAL_SIGNOFF_COMPLETE=0

PROPOSAL_SOURCE="/home/servermtsn2kolut/.hermes/document_cache/doc_2954fa0c7a5c_Proposal_Sistem_CBT_MTsN2_Kolaka_Utara.docx"
GAP_AUDIT_DOC="docs/cbt-proposal-gap-audit.md"
FINAL_EVIDENCE_DOC="docs/cbt-release-final-evidence.md"
EVIDENCE_TEMPLATE_DOC="docs/cbt-release-evidence-template.md"
DEVICE_MATRIX_DOC="apps/mobile/DEVICE_TEST_MATRIX.md"
SCRIPT_DOC="deploy/scripts/cbt-final-readiness.sh"

GAP_AUDIT_JSON=""
FINAL_EVIDENCE_JSON=""
FINAL_SIGNOFF_JSON=""
FINAL_READINESS_MD=""
LOG_DIR=""
SECRET_SCAN_STATUS="pending"
SECRET_SCAN_FINDINGS=()
SECRET_SCAN_FILES=()

AUTOMATED_NAMES=()
AUTOMATED_STATUSES=()
AUTOMATED_DETAILS=()
AUTOMATED_LOGS=()

HEALTH_NAMES=()
HEALTH_STATUSES=()
HEALTH_DETAILS=()
HEALTH_LOGS=()

usage() {
	cat <<'USAGE'
Usage: deploy/scripts/cbt-final-readiness.sh [options]

Generate read-only final CBT readiness evidence. By default this writes only to
tmp/cbt-final-readiness inside the repository.

Options:
  --output <dir>                         Write evidence to a caller-provided directory.
  --run-ops-health                       Run read-only `make ops-health` and capture status.
  --manual-device-matrix-complete        Mark physical Android device matrix evidence complete.
  --manual-operator-rehearsal-complete   Mark operator rehearsal evidence complete.
  --manual-final-signoff-complete        Mark final operator sign-off evidence complete.
  -h, --help                             Show this help.

The script does not deploy, change PM2 process state, run migrations, mutate SQL state, or
change product runtime. It writes generated evidence only.
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
		--run-ops-health)
			RUN_OPS_HEALTH=1
			shift
			;;
		--manual-device-matrix-complete)
			MANUAL_DEVICE_MATRIX_COMPLETE=1
			shift
			;;
		--manual-operator-rehearsal-complete)
			MANUAL_OPERATOR_REHEARSAL_COMPLETE=1
			shift
			;;
		--manual-final-signoff-complete)
			MANUAL_FINAL_SIGNOFF_COMPLETE=1
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
GAP_AUDIT_JSON="${OUTPUT_DIR}/cbt-proposal-gap-audit.json"
FINAL_EVIDENCE_JSON="${OUTPUT_DIR}/cbt-final-evidence.json"
FINAL_SIGNOFF_JSON="${OUTPUT_DIR}/cbt-final-signoff.json"
FINAL_READINESS_MD="${OUTPUT_DIR}/cbt-final-readiness.md"

if [ -e "$OUTPUT_DIR" ] && [ ! -d "$OUTPUT_DIR" ]; then
	printf 'ERROR: output path exists but is not a directory: %s\n' "$OUTPUT_DIR" >&2
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
	if [[ "$path" == "$OUTPUT_DIR/"* ]]; then
		printf '%s\n' "${path#"$OUTPUT_DIR/"}"
	else
		printf '%s\n' "$path"
	fi
}

record_automated() {
	AUTOMATED_NAMES+=("$1")
	AUTOMATED_STATUSES+=("$2")
	AUTOMATED_DETAILS+=("$3")
	AUTOMATED_LOGS+=("${4:-}")
}

record_health() {
	HEALTH_NAMES+=("$1")
	HEALTH_STATUSES+=("$2")
	HEALTH_DETAILS+=("$3")
	HEALTH_LOGS+=("${4:-}")
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
		record_automated "$name" "pass" "$display" "$(relative_output_path "$log_path")"
	else
		record_automated "$name" "fail" "$display exited $status" "$(relative_output_path "$log_path")"
	fi
}

run_health_command() {
	local name="make ops-health"
	local tmp_log="${LOG_DIR}/ops-health.tmp"
	local log_path="${LOG_DIR}/ops-health.log"
	local status=0

	if [ "$RUN_OPS_HEALTH" -ne 1 ]; then
		record_health "$name" "skipped" "not requested; run with --run-ops-health for read-only live health status" ""
		return
	fi

	set +e
	(
		cd "$REPO_ROOT"
		GIT_OPTIONAL_LOCKS=0 make ops-health
	) > "$tmp_log" 2>&1
	status=$?
	set -e

	redact_stream < "$tmp_log" > "$log_path"
	rm -f "$tmp_log"

	if [ "$status" -eq 0 ]; then
		record_health "$name" "pass" "read-only health command completed" "$(relative_output_path "$log_path")"
	else
		record_health "$name" "fail" "read-only health command exited $status" "$(relative_output_path "$log_path")"
	fi
}

doc_exists_json_entry() {
	local key="$1"
	local doc_path="$2"
	local exists="false"
	if [ -f "${REPO_ROOT}/${doc_path}" ]; then
		exists="true"
	fi
	printf '    "%s": { "path": "%s", "exists": %s }' "$(json_escape "$key")" "$(json_escape "$doc_path")" "$exists"
}

all_required_docs_exist() {
	for doc_path in "$GAP_AUDIT_DOC" "$FINAL_EVIDENCE_DOC" "$EVIDENCE_TEMPLATE_DOC" "$DEVICE_MATRIX_DOC" "$SCRIPT_DOC"; do
		[ -f "${REPO_ROOT}/${doc_path}" ] || return 1
	done
	return 0
}

manual_status() {
	if [ "$1" -eq 1 ]; then
		printf 'complete_by_operator_flag'
	else
		printf 'pending_manual_evidence'
	fi
}

final_signoff_status() {
	if [ "$MANUAL_FINAL_SIGNOFF_COMPLETE" -eq 1 ]; then
		printf 'complete_by_operator_flag'
	else
		printf 'pending_manual_signoff'
	fi
}

go_no_go_status() {
	if [ "$MANUAL_DEVICE_MATRIX_COMPLETE" -eq 1 ] && \
		[ "$MANUAL_OPERATOR_REHEARSAL_COMPLETE" -eq 1 ] && \
		[ "$MANUAL_FINAL_SIGNOFF_COMPLETE" -eq 1 ]; then
		printf 'ready_for_rehearsal'
	else
		printf 'pending_manual_signoff'
	fi
}

write_gap_audit_json() {
	{
		printf '{\n'
		printf '  "generated_at": "%s",\n' "$(json_escape "$GENERATED_AT")"
		printf '  "repo_root": "%s",\n' "$(json_escape "$REPO_ROOT")"
		printf '  "commit": "%s",\n' "$(json_escape "$HEAD_SHA")"
		printf '  "proposal_source": "%s",\n' "$(json_escape "$PROPOSAL_SOURCE")"
		printf '  "status_vocabulary": ["Implemented", "Partial", "Planned manual evidence", "Out of scope adapted"],\n'
		printf '  "matrix": [\n'
		write_matrix_row "Multi-mode assessment" "4.1 Multi-Mode Assessment Engine" "Partial" "CBT sessions/events, Bank Soal, paket, non-test assessments, results hubs, and Flutter exam runtime exist." "Keep CBT and non-test assessment split; add adaptive/leaderboard/unlimited-attempt behavior only after explicit product decision." ","
		write_matrix_row "Question types" "4.2 Tipe Soal yang Didukung" "Partial" "Pilihan ganda, essay, rich content, assets, import/export, and manual essay grading are present." "Treat additional item types as backlog with schema/API/rendering/scoring tests." ","
		write_matrix_row "Anti-cheat layers" "4.3 Sistem Anti-Cheat Berlapis" "Partial" "Strong exam tokens, device telemetry, FLAG_SECURE, app-switch/resume events, heartbeat, degraded-mode submit guard, and proctor warnings are present." "Continue BYOD-realistic deterrence and collect physical-device evidence; do not claim full kiosk control." ","
		write_matrix_row "Room management" "4.4 Room Management" "Implemented" "Sessions, rooms, seats, room proctors, readiness, print packs, handover, and participant operations exist." "Verify room setup in operator rehearsal and record QR/access-code gaps as backlog." ","
		write_matrix_row "Proctor dashboard" "4.5 Dashboard Pengawas" "Partial" "Room monitoring, warnings/events, force submit, reset access, participant flagging, print packs, and operational recaps exist." "Capture proctor evidence during rehearsal for status, warnings, pending answers, device mismatch, and submit guard." ","
		write_matrix_row "Audit trail" "4.5.3 Audit Trail & Evidensi, 5.3 Audit Log" "Partial" "Audit middleware, CBT event telemetry, auth audit events, session/room audit routes, and evidence templates exist." "Use generated readiness/evidence plus ops archive; hash-chain audit needs separate design." ","
		write_matrix_row "Analytics" "8.1 Analitik Butir Soal" "Partial" "Results hubs, scoring, item-analysis route, essay grading, and grade sync foundations exist." "Keep item-analysis evidence in post-exam review and expand metrics with validated formulas/tests." ","
		write_matrix_row "Reports" "8.2 Laporan yang Tersedia" "Partial" "CSV exports, exam cards, minutes, operational recap, session/event results, and evidence templates exist." "Use current artifacts for release; track PDF/Excel executive reports as backlog." ","
		write_matrix_row "ISO controls" "5.3, 6.1, 6.2, 6.3" "Partial" "RBAC, refresh sessions, auth audit, rate limiting, upload hygiene, generic 500 hygiene, docs guard, checksums, and secret-scan evidence exist." "Record ISO as control alignment, not certification; attach manual security/ops evidence before external claims." ","
		write_matrix_row "Infrastructure" "5.1 Stack Teknologi, 5.2 Arsitektur Sistem, 9.2 Kebutuhan Infrastruktur" "Out of scope adapted" "Approved runtime is 3 VPS targets with SvelteKit, Go, PostgreSQL, PUSAKA worker, Flutter APK, native PM2." "Preserve service boundaries and gather read-only ops-health evidence; do not introduce PocketBase/runtime SQLite." ","
		write_matrix_row "Risks" "11 Manajemen Risiko" "Planned manual evidence" "Some mitigations exist: server-side timer, pending answers, heartbeat, restore, auth hardening, runbooks, and audit templates." "Complete device matrix, operator rehearsal, evidence review, and final go/no-go sign-off before production acceptance." ""
		printf '  ],\n'
		printf '  "boundary": {\n'
		printf '    "public_api_cbt_route_tree_required": false,\n'
		printf '    "student_runtime_api": "/api/exam/*"\n'
		printf '  },\n'
		printf '  "secret_scan": { "status": "%s" }\n' "$(json_escape "$SECRET_SCAN_STATUS")"
		printf '}\n'
	} > "$GAP_AUDIT_JSON"
}

write_matrix_row() {
	local area="$1"
	local section="$2"
	local status="$3"
	local evidence="$4"
	local next_action="$5"
	local comma="$6"

	printf '    {\n'
	printf '      "area": "%s",\n' "$(json_escape "$area")"
	printf '      "proposal_section": "%s",\n' "$(json_escape "$section")"
	printf '      "status": "%s",\n' "$(json_escape "$status")"
	printf '      "evidence": "%s",\n' "$(json_escape "$evidence")"
	printf '      "next_action": "%s"\n' "$(json_escape "$next_action")"
	if [ -n "$comma" ]; then
		printf '    },\n'
	else
		printf '    }\n'
	fi
}

write_final_evidence_json() {
	local device_status=""
	local operator_status=""
	local signoff_status=""
	device_status="$(manual_status "$MANUAL_DEVICE_MATRIX_COMPLETE")"
	operator_status="$(manual_status "$MANUAL_OPERATOR_REHEARSAL_COMPLETE")"
	signoff_status="$(final_signoff_status)"

	{
		printf '{\n'
		printf '  "generated_at": "%s",\n' "$(json_escape "$GENERATED_AT")"
		printf '  "repo_root": "%s",\n' "$(json_escape "$REPO_ROOT")"
		printf '  "output_dir": "%s",\n' "$(json_escape "$OUTPUT_DIR")"
		printf '  "branch": "%s",\n' "$(json_escape "$BRANCH_NAME")"
		printf '  "commit": "%s",\n' "$(json_escape "$HEAD_SHA")"
		printf '  "docs": {\n'
		doc_exists_json_entry "proposal_gap_audit" "$GAP_AUDIT_DOC"; printf ',\n'
		doc_exists_json_entry "final_release_evidence" "$FINAL_EVIDENCE_DOC"; printf ',\n'
		doc_exists_json_entry "release_evidence_template" "$EVIDENCE_TEMPLATE_DOC"; printf ',\n'
		doc_exists_json_entry "device_test_matrix" "$DEVICE_MATRIX_DOC"; printf ',\n'
		doc_exists_json_entry "final_readiness_script" "$SCRIPT_DOC"; printf '\n'
		printf '  },\n'
		printf '  "automated_evidence": [\n'
		for i in "${!AUTOMATED_NAMES[@]}"; do
			printf '    { "name": "%s", "status": "%s", "detail": "%s", "log": "%s" }' \
				"$(json_escape "${AUTOMATED_NAMES[$i]}")" \
				"$(json_escape "${AUTOMATED_STATUSES[$i]}")" \
				"$(json_escape "${AUTOMATED_DETAILS[$i]}")" \
				"$(json_escape "${AUTOMATED_LOGS[$i]}")"
			if [ "$i" -eq "$((${#AUTOMATED_NAMES[@]} - 1))" ]; then
				printf '\n'
			else
				printf ',\n'
			fi
		done
		printf '  ],\n'
		printf '  "manual_evidence": {\n'
		printf '    "device_matrix": {\n'
		printf '      "status": "%s",\n' "$(json_escape "$device_status")"
		printf '      "path": "%s",\n' "$(json_escape "$DEVICE_MATRIX_DOC")"
		printf '      "requires_physical_android_devices": true,\n'
		printf '      "minimum_vendors": 2\n'
		printf '    },\n'
		printf '    "operator_rehearsal": {\n'
		printf '      "status": "%s",\n' "$(json_escape "$operator_status")"
		printf '      "requires_live_operator_rehearsal": true,\n'
		printf '      "flow": "Bank Soal to Asesmen Persiapan to Pelaksanaan/Pengawasan to Flutter APK to Hasil/Post-exam review"\n'
		printf '    },\n'
		printf '    "final_signoff": { "status": "%s" }\n' "$(json_escape "$signoff_status")"
		printf '  },\n'
		printf '  "health_commands": [\n'
		for i in "${!HEALTH_NAMES[@]}"; do
			printf '    { "name": "%s", "status": "%s", "detail": "%s", "log": "%s" }' \
				"$(json_escape "${HEALTH_NAMES[$i]}")" \
				"$(json_escape "${HEALTH_STATUSES[$i]}")" \
				"$(json_escape "${HEALTH_DETAILS[$i]}")" \
				"$(json_escape "${HEALTH_LOGS[$i]}")"
			if [ "$i" -eq "$((${#HEALTH_NAMES[@]} - 1))" ]; then
				printf '\n'
			else
				printf ',\n'
			fi
		done
		printf '  ],\n'
		write_tests_manifest_json
		printf ',\n'
		printf '  "secret_scan": { "status": "%s" }\n' "$(json_escape "$SECRET_SCAN_STATUS")"
		printf '}\n'
	} > "$FINAL_EVIDENCE_JSON"
}

write_tests_manifest_json() {
	cat <<'JSON'
  "tests_manifest": [
    {
      "name": "diff whitespace",
      "command": "git diff --check",
      "scope": "repository whitespace validation"
    },
    {
      "name": "targeted docs guard",
      "command": "cd apps/web-admin && npm run test:unit -- src/lib/cbt/proposal-integration-docs.test.ts",
      "scope": "CBT proposal/final readiness docs and script guard"
    },
    {
      "name": "web check",
      "command": "cd apps/web-admin && npm run check",
      "scope": "SvelteKit accessibility and types"
    },
    {
      "name": "full web tests",
      "command": "cd apps/web-admin && npm run test:unit",
      "scope": "web-admin Vitest suite"
    },
    {
      "name": "go test",
      "command": "cd services/core-api && go test ./...",
      "scope": "Core API tests"
    },
    {
      "name": "go build",
      "command": "cd services/core-api && go build -o /dev/null ./cmd/api",
      "scope": "Core API build"
    },
    {
      "name": "flutter analyze",
      "command": "cd apps/mobile && /home/servermtsn2kolut/development/flutter/bin/flutter analyze",
      "scope": "Flutter static analysis"
    },
    {
      "name": "flutter test",
      "command": "cd apps/mobile && /home/servermtsn2kolut/development/flutter/bin/flutter test",
      "scope": "Flutter tests"
    },
    {
      "name": "ops health",
      "command": "make ops-health",
      "scope": "read-only ops health status"
    }
  ]
JSON
}

write_final_signoff_json() {
	local go_no_go=""
	go_no_go="$(go_no_go_status)"

	{
		printf '{\n'
		printf '  "generated_at": "%s",\n' "$(json_escape "$GENERATED_AT")"
		printf '  "commit": "%s",\n' "$(json_escape "$HEAD_SHA")"
		printf '  "go_no_go": "%s",\n' "$(json_escape "$go_no_go")"
		printf '  "production_go": false,\n'
		printf '  "manual_inputs": {\n'
		printf '    "device_matrix_complete": %s,\n' "$(bool_json "$MANUAL_DEVICE_MATRIX_COMPLETE")"
		printf '    "operator_rehearsal_complete": %s,\n' "$(bool_json "$MANUAL_OPERATOR_REHEARSAL_COMPLETE")"
		printf '    "final_signoff_complete": %s\n' "$(bool_json "$MANUAL_FINAL_SIGNOFF_COMPLETE")"
		printf '  },\n'
		printf '  "blockers": [\n'
		write_blockers_json
		printf '  ],\n'
		printf '  "boundary": {\n'
		printf '    "public_api_cbt_route_tree_required": false,\n'
		printf '    "student_runtime_api": "/api/exam/*"\n'
		printf '  },\n'
		printf '  "secret_scan": { "status": "%s" }\n' "$(json_escape "$SECRET_SCAN_STATUS")"
		printf '}\n'
	} > "$FINAL_SIGNOFF_JSON"
}

bool_json() {
	if [ "$1" -eq 1 ]; then
		printf 'true'
	else
		printf 'false'
	fi
}

write_blockers_json() {
	local blockers=()
	if [ "$MANUAL_DEVICE_MATRIX_COMPLETE" -ne 1 ]; then
		blockers+=("manual device matrix evidence pending")
	fi
	if [ "$MANUAL_OPERATOR_REHEARSAL_COMPLETE" -ne 1 ]; then
		blockers+=("operator rehearsal evidence pending")
	fi
	if [ "$MANUAL_FINAL_SIGNOFF_COMPLETE" -ne 1 ]; then
		blockers+=("final sign-off pending")
	fi

	for i in "${!blockers[@]}"; do
		printf '    "%s"' "$(json_escape "${blockers[$i]}")"
		if [ "$i" -eq "$((${#blockers[@]} - 1))" ]; then
			printf '\n'
		else
			printf ',\n'
		fi
	done
}

write_markdown() {
	local go_no_go=""
	local device_status=""
	local operator_status=""
	local signoff_status=""
	go_no_go="$(go_no_go_status)"
	device_status="$(manual_status "$MANUAL_DEVICE_MATRIX_COMPLETE")"
	operator_status="$(manual_status "$MANUAL_OPERATOR_REHEARSAL_COMPLETE")"
	signoff_status="$(final_signoff_status)"

	{
		printf '# CBT Final Readiness Evidence\n\n'
		printf '%s\n' "- Generated at: \`${GENERATED_AT}\`"
		printf '%s\n' "- Repository: \`${REPO_ROOT}\`"
		printf '%s\n' "- Branch: \`${BRANCH_NAME}\`"
		printf '%s\n' "- Commit: \`${HEAD_SHA}\`"
		printf '%s\n' "- Go/no-go: \`${go_no_go}\`"
		printf '%s\n' "- Production go claimed: \`false\`"
		printf '%s\n\n' "- Output directory: \`${OUTPUT_DIR}\`"

		printf '## Automated Evidence\n\n'
		printf '| Item | Status | Detail | Log |\n'
		printf '|------|--------|--------|-----|\n'
		for i in "${!AUTOMATED_NAMES[@]}"; do
			printf '| `%s` | `%s` | %s | %s |\n' \
				"${AUTOMATED_NAMES[$i]}" \
				"${AUTOMATED_STATUSES[$i]}" \
				"${AUTOMATED_DETAILS[$i]//|/\\|}" \
				"${AUTOMATED_LOGS[$i]}"
		done

		printf '\n## Manual Evidence\n\n'
		printf '| Evidence | Status | Requirement |\n'
		printf '|----------|--------|-------------|\n'
		printf '| Device matrix | `%s` | physical Android devices, minimum two vendors |\n' "$device_status"
		printf '| Operator rehearsal | `%s` | live operator/pengawas flow and proctor evidence |\n' "$operator_status"
		printf '| Final sign-off | `%s` | operator decision with reviewer and rollback owner |\n' "$signoff_status"

		printf '\n## Health Commands\n\n'
		printf '| Command | Status | Detail | Log |\n'
		printf '|---------|--------|--------|-----|\n'
		for i in "${!HEALTH_NAMES[@]}"; do
			printf '| `%s` | `%s` | %s | %s |\n' \
				"${HEALTH_NAMES[$i]}" \
				"${HEALTH_STATUSES[$i]}" \
				"${HEALTH_DETAILS[$i]//|/\\|}" \
				"${HEALTH_LOGS[$i]}"
		done

		printf '\n## Generated Artifacts\n\n'
		printf '%s\n' "- \`cbt-proposal-gap-audit.json\`"
		printf '%s\n' "- \`cbt-final-evidence.json\`"
		printf '%s\n' "- \`cbt-final-signoff.json\`"
		printf '%s\n\n' "- \`cbt-final-readiness.md\`"

		printf '## Boundary\n\n'
		printf '%s\n' '- Tidak deploy, tidak mengubah proses PM2, tidak migration, tidak live DB write, dan tidak live exam mutation.'
		printf '%s\n' '- Tidak membuat public SvelteKit route tree `/api/cbt/**` baru.'
		printf '%s\n' '- Flutter tetap memakai `services/core-api` `/api/exam/*` untuk runtime siswa.'
		printf '%s\n' '- This host can generate automated evidence, but physical Android devices and operator rehearsal remain manual evidence until recorded.'
	} > "$FINAL_READINESS_MD"
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
	local key_pattern="[\"']?(password|secret|api[ _-]?key|answer[ _-]?key)[\"']?[[:space:]]*[:=][[:space:]]*[\"']?[^\"'[:space:],}]+"

	while IFS= read -r file; do
		[ -n "$file" ] || continue
		rel_path="$(relative_output_path "$file")"
		SECRET_SCAN_FILES+=("$rel_path")
		scan_secret_pattern "$file" "$rel_path" "authorization-bearer" "$auth_pattern"
		scan_secret_pattern "$file" "$rel_path" "sensitive-assignment" "$key_pattern"
	done < <(find "$OUTPUT_DIR" -type f \( -name '*.md' -o -name '*.json' -o -name '*.log' \) -print | LC_ALL=C sort)

	if [ "${#SECRET_SCAN_FINDINGS[@]}" -gt 0 ]; then
		SECRET_SCAN_STATUS="fail"
		printf 'ERROR: final readiness secret scan failed:\n' >&2
		local finding=""
		for finding in "${SECRET_SCAN_FINDINGS[@]}"; do
			printf '  %s\n' "$finding" >&2
		done
		return 1
	fi

	SECRET_SCAN_STATUS="pass"
	return 0
}

GENERATED_AT="$(date -Iseconds)"
HEAD_SHA="$(GIT_OPTIONAL_LOCKS=0 git -C "$REPO_ROOT" rev-parse HEAD 2>/dev/null || true)"
BRANCH_NAME="$(GIT_OPTIONAL_LOCKS=0 git -C "$REPO_ROOT" branch --show-current 2>/dev/null || true)"

if [ -z "$HEAD_SHA" ]; then
	HEAD_SHA="unknown"
fi
if [ -z "$BRANCH_NAME" ]; then
	BRANCH_NAME="unknown"
fi

run_logged_command "git-diff-check" "$REPO_ROOT" "git diff --check" git diff --check

if all_required_docs_exist; then
	record_automated "docs-existence" "pass" "final audit, final evidence, template, device matrix, and script docs exist" ""
else
	record_automated "docs-existence" "fail" "one or more final audit/readiness docs are missing" ""
fi

record_automated "tests-manifest" "recorded" "validation command manifest written to cbt-final-evidence.json" ""
run_health_command

write_gap_audit_json
write_final_evidence_json
write_final_signoff_json
write_markdown

if scan_generated_evidence_for_secrets; then
	SECRET_SCAN_STATUS="pass"
else
	SECRET_SCAN_STATUS="fail"
fi

write_gap_audit_json
write_final_evidence_json
write_final_signoff_json
write_markdown

printf 'CBT final readiness evidence written:\n'
printf '  Gap audit JSON: %s\n' "$GAP_AUDIT_JSON"
printf '  Evidence JSON:  %s\n' "$FINAL_EVIDENCE_JSON"
printf '  Signoff JSON:   %s\n' "$FINAL_SIGNOFF_JSON"
printf '  Markdown:       %s\n' "$FINAL_READINESS_MD"

if [ "$SECRET_SCAN_STATUS" = "fail" ]; then
	exit 1
fi
