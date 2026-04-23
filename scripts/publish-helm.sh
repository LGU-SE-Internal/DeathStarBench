#!/bin/bash
# Package all DeathStarBench helm charts and build a combined index for gh-pages.
#
# Usage: bash scripts/publish-helm.sh
#
# Produces .deploy/*.tgz + .deploy/index.yaml that can be served at
# https://<org>.github.io/<repo>/ as a helm repo.
set -euo pipefail

CHART_DIRS=(
  "hotelReservation/helm-chart/hotelreservation"
  "socialNetwork/helm-chart/socialnetwork"
  "mediaMicroservices/helm-chart/mediamicroservices"
)

if [ -z "${GITHUB_REPOSITORY:-}" ]; then
    GIT_REMOTE=$(git config --get remote.origin.url 2>/dev/null || echo "")
    if [[ $GIT_REMOTE =~ github.com[:/]([^/]+)/([^/.]+) ]]; then
        ORG_NAME="${BASH_REMATCH[1]}"
        REPO_NAME="${BASH_REMATCH[2]}"
    else
        ORG_NAME="lgu-se-internal"
        REPO_NAME="DeathStarBench"
        echo "Warning: Falling back to default ORG_NAME=$ORG_NAME REPO_NAME=$REPO_NAME"
    fi
else
    ORG_NAME=$(echo "$GITHUB_REPOSITORY" | cut -d'/' -f1 | tr '[:upper:]' '[:lower:]')
    REPO_NAME=$(echo "$GITHUB_REPOSITORY" | cut -d'/' -f2)
fi

REPO_URL="https://${ORG_NAME}.github.io/${REPO_NAME}"
echo "Target helm repo URL: $REPO_URL"

mkdir -p .deploy

for dir in "${CHART_DIRS[@]}"; do
    if [ ! -f "$dir/Chart.yaml" ]; then
        echo "ERROR: Chart.yaml not found in $dir" >&2
        exit 1
    fi
    echo "Packaging chart: $dir"
    # Best-effort dependency update; skip silently if no dependencies / no network.
    helm dependency update "$dir" || echo "  (helm dependency update failed, continuing)"
    helm package "$dir" -d .deploy
done

if [ -f .deploy/index.yaml ]; then
    helm repo index .deploy --url "$REPO_URL" --merge .deploy/index.yaml
else
    helm repo index .deploy --url "$REPO_URL"
fi

echo
echo "Wrote .deploy/ contents:"
ls -l .deploy
