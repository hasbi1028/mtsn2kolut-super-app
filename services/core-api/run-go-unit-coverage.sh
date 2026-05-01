#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

coverage_dir="${COVERAGE_DIR:-${TMPDIR:-/tmp}/mtsn2kolut-core-api-coverage}"
unit_profile="${COVERAGE_PROFILE:-${coverage_dir}/coverage.unit.out}"
threshold="${COVERAGE_THRESHOLD:-}"

mkdir -p "$coverage_dir"

packages=()
if ! package_list="$(go list ./...)"; then
	echo "failed to list Go packages for unit coverage" >&2
	exit 1
fi

while IFS= read -r package; do
	case "$package" in
		*/cmd/api | */internal/repository/postgres)
			continue
			;;
	esac
	packages+=("$package")
done <<< "$package_list"

if [[ ${#packages[@]} -eq 0 ]]; then
	echo "no packages selected for unit coverage" >&2
	exit 1
fi

go test "${packages[@]}" -covermode=atomic -coverprofile="$unit_profile"

coverage_output="$(go tool cover -func="$unit_profile")"
if [[ "${COVERAGE_VERBOSE:-}" == "1" ]]; then
	printf '%s\n' "$coverage_output"
else
	printf '%s\n' "$coverage_output" | awk '/^total:/ { print }'
fi

coverage_pct="$(printf '%s\n' "$coverage_output" | awk '/^total:/ { gsub(/%/, "", $3); print $3 }')"
if [[ -z "$coverage_pct" ]]; then
	echo "could not read total coverage from ${unit_profile}" >&2
	exit 1
fi

if [[ -n "$threshold" ]]; then
	awk -v got="$coverage_pct" -v want="$threshold" 'BEGIN { exit !(got + 0 >= want + 0) }' || {
		echo "coverage ${coverage_pct}% is below threshold ${threshold}%" >&2
		exit 1
	}
fi

echo "unit coverage profile: ${unit_profile}"
