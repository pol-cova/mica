#!/bin/sh
set -eu

cache_dir="${GOCACHE:-/tmp/mica-go-cache}"

run_check() {
	label="$1"
	package="$2"
	test_name="$3"
	if GOCACHE="$cache_dir" go test "$package" -run "^${test_name}$" -count=1 >/dev/null; then
		printf 'PASS  %s\n' "$label"
	else
		printf 'FAIL  %s\n' "$label"
		exit 1
	fi
}

printf 'Mica MVP evaluation\n\n'
run_check "detects a telemetry regression" ./internal/prometheus TestCompareUsesRangeWindowsAndCalculatesDegradation
run_check "preserves incident evidence and workflow" ./internal/mcpserver TestToolsListAndIncidentWorkflow
run_check "requires evidence and human review for actions" ./internal/incidents TestProposalRequiresIncidentEvidenceAndCanBeReviewed
run_check "requires explicit approval for delivery" ./internal/incidents TestPreparedUpdateRequiresExplicitApproverForDelivery
run_check "verifies recovery against saved signals" ./internal/incidents TestRecoveryUsesSavedIncidentSignals
printf '\n5/5 checks passed\n'
