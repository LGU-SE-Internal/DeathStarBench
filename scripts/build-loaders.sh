#!/bin/bash
# Build the three DeathStarBench wrk2 loader images.
#
# MUST be run from the repo root: the loader Dockerfiles reference
# `../wrk2` which only resolves when the docker build context is the
# repository root.
#
# Env:
#   REGISTRY  image registry/prefix (default: deathstarbench)
#   TAG       image tag             (default: latest)
#   PUSH      if set to 1, push images after building
#
# Example:
#   REGISTRY=10.10.10.240/library TAG=dev PUSH=1 bash scripts/build-loaders.sh
set -euo pipefail

REGISTRY="${REGISTRY:-deathstarbench}"
TAG="${TAG:-latest}"
PUSH="${PUSH:-0}"

# Ensure we're at repo root (must see the wrk2 directory).
if [ ! -d "wrk2" ]; then
    echo "ERROR: run from repo root (expected ./wrk2 to exist)" >&2
    exit 1
fi

# Initialize the luajit submodule if missing; wrk2 build requires it.
if [ ! -f "wrk2/deps/luajit/Makefile" ]; then
    echo "Initializing wrk2/deps/luajit submodule..."
    git submodule update --init --recursive
fi

declare -A LOADERS=(
    [hotelreservation-loader]="hotelReservation/Dockerfile-loader"
    [socialnetwork-loader]="socialNetwork/Dockerfile-loader"
    [mediamicroservices-loader]="mediaMicroservices/Dockerfile-loader"
)

for name in "${!LOADERS[@]}"; do
    dockerfile="${LOADERS[$name]}"
    image="${REGISTRY}/${name}:${TAG}"
    echo "==> docker build -f ${dockerfile} -t ${image} ."
    docker build -f "${dockerfile}" -t "${image}" .
    if [ "${PUSH}" = "1" ]; then
        echo "==> docker push ${image}"
        docker push "${image}"
    fi
done

echo "Done."
