# loader-datasets image

Pre-baked read-only image containing the dataset files and helper scripts
the `socialNetwork` and `mediaMicroservices` load-generator deployments
and data-init Jobs need at runtime.

## Why

The previous `fetch-datasets` init container ran
`git clone https://github.com/delimitrou/DeathStarBench.git` (or the LGU
fork) into an EmptyDir volume on every pod restart. On clusters with
constrained or flaky GitHub egress, that clone times out, half-completes,
or hangs — the init container CrashLoopBackOffs and the load-generator
never starts, silently dropping fault-injection runs that depend on
sustained traffic.

This image bakes the dataset content directly so the init container only
needs to `cp -r /datasets/. /shared/` from a registry the cluster can
already pull from.

## Layout

```
/datasets/socialNetwork/datasets/social-graph/{ego-twitter,socfb-Reed98,soc-twitter-follows-mun}/...
/datasets/socialNetwork/scripts/init_social_graph.py
/datasets/mediaMicroservices/datasets/tmdb/{casts.json,movies.json,...}
/datasets/mediaMicroservices/scripts/write_movie_info.py
```

## Build locally

From the repository root (NOT this directory):

```bash
docker build -f loader-datasets/Dockerfile \
  -t deathstarbench/loader-datasets:latest .
```

## CI publish

`.github/workflows/loader-datasets-push.yaml` builds and pushes
`deathstarbench/loader-datasets:<version>` on every push to `master`
that touches:
- `loader-datasets/**`
- `socialNetwork/datasets/**`
- `socialNetwork/scripts/**`
- `mediaMicroservices/datasets/**`
- `mediaMicroservices/scripts/**`

It uses the same `DOCKER_REGISTRY_LOGIN` / `DOCKER_REGISTRY_PASSWORD`
secrets the other DSB image push workflows use.
