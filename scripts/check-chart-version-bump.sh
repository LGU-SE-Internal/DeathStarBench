#!/usr/bin/env bash
# Fail if a helm chart's tree changed between BASE_SHA and HEAD_SHA but
# its Chart.yaml `version` field was not bumped. Designed for PR CI, but
# also runnable locally:
#
#   bash scripts/check-chart-version-bump.sh origin/master HEAD
#
# Auto-discovers every `*/helm-chart/*/Chart.yaml` as a guarded chart.
# A chart "changed" when any file under its chart root differs between
# the two revs. To override for genuinely non-substantive changes
# (e.g. doc-only edits), put `[skip-chart-bump]` in a commit message
# somewhere in the PR's commit range.
set -euo pipefail

BASE="${1:-}"
HEAD="${2:-HEAD}"

if [ -z "$BASE" ]; then
  echo "usage: $0 <base-sha-or-ref> [head-sha-or-ref]" >&2
  exit 2
fi

mapfile -t CHARTS < <(git ls-tree -r --name-only "$HEAD" | grep -E '/helm-chart/[^/]+/Chart\.yaml$' | sort -u)

if [ ${#CHARTS[@]} -eq 0 ]; then
  echo "No helm charts discovered under */helm-chart/*/Chart.yaml; nothing to check."
  exit 0
fi

get_version() {
  local ref="$1" path="$2"
  git show "$ref:$path" 2>/dev/null | awk -F': *' '/^version:/ {print $2; exit}' | tr -d '"[:space:]'
}

fail=0
for chart_yaml in "${CHARTS[@]}"; do
  chart_root="${chart_yaml%/Chart.yaml}"
  chart_name=$(basename "$chart_root")

  if ! git diff --quiet "$BASE" "$HEAD" -- "$chart_root"; then
    base_version=$(get_version "$BASE" "$chart_yaml" || true)
    head_version=$(get_version "$HEAD" "$chart_yaml")

    if [ -z "$head_version" ]; then
      echo "::error file=$chart_yaml::Chart.yaml missing 'version:' field"
      fail=1
      continue
    fi

    if [ -z "$base_version" ]; then
      echo "ok: $chart_name is new (version=$head_version)"
      continue
    fi

    if [ "$base_version" = "$head_version" ]; then
      if git log --format=%B "$BASE..$HEAD" -- "$chart_root" | grep -qF '[skip-chart-bump]'; then
        echo "ok: $chart_name changed but commit message has [skip-chart-bump] (version=$head_version)"
        continue
      fi
      changed=$(git diff --name-only "$BASE" "$HEAD" -- "$chart_root")
      echo "::error file=$chart_yaml::Chart '$chart_name' changed but version was not bumped (still $head_version). Edit $chart_yaml and bump 'version' (e.g. $head_version -> next patch) so helm repo caches invalidate. Include [skip-chart-bump] in a commit message to override for genuinely non-substantive changes."
      echo "  changed paths:"
      printf '    %s\n' $changed
      fail=1
    else
      echo "ok: $chart_name bumped $base_version -> $head_version"
    fi
  fi
done

if [ $fail -ne 0 ]; then
  echo
  echo "One or more charts were edited without a version bump. Fix the Chart.yaml 'version' field for each flagged chart and push again." >&2
  exit 1
fi
