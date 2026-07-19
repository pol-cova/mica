#!/usr/bin/env sh
set -eu

root=".agents/skills"
failed=0

check() {
  name="$1"
  pattern="$2"
  file="$3"
  if grep -qi "$pattern" "$file"; then
    printf 'PASS %-29s %s\n' "$name" "$file"
  else
    printf 'FAIL %-29s %s\n' "$name" "$file"
    failed=1
  fi
}

for file in "$root"/*/SKILL.md; do
  [ -f "$file" ] || continue
  check "skill selection" '^name:' "$file"
  check "evidence gate" 'evidence\|baseline\|quantitative' "$file"
  check "safety boundary" 'never\|do not\|stop' "$file"
  check "final outcome" 'final response\|return\|output\|structure' "$file"
done

if [ "$failed" -ne 0 ]; then
  exit 1
fi
