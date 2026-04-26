# DeathStarBench loader datasets image.
#
# A tiny read-only image that ships only the dataset and init-script files
# the load-generator and data-init-job containers need at runtime. Replaces
# the previous `git clone https://github.com/.../DeathStarBench.git` init
# container, which was flaky on clusters with constrained GitHub egress.
#
# Layout under /datasets/ (consumed by `cp -r` in the chart init container):
#
#   /datasets/socialNetwork/datasets/social-graph/...
#   /datasets/socialNetwork/scripts/init_social_graph.py (+ helpers)
#   /datasets/mediaMicroservices/datasets/tmdb/{casts.json,movies.json,...}
#   /datasets/mediaMicroservices/scripts/write_movie_info.py (+ helpers)
#
# Build context MUST be the repository root so that ./socialNetwork and
# ./mediaMicroservices resolve, matching the convention of the loader images:
#
#   docker build -f loader-datasets/Dockerfile -t deathstarbench/loader-datasets:latest .
#
# The image's filesystem is read-only at runtime; the init container in each
# chart copies /datasets/. into the shared EmptyDir volume the data-init and
# loadgen containers already mount.

FROM busybox:1.36

# socialNetwork: social-graph datasets + init scripts (used by data-init-job
# and load-generator deployment to seed users / follows in the running
# socialNetwork instance).
COPY ./socialNetwork/datasets   /datasets/socialNetwork/datasets
COPY ./socialNetwork/scripts    /datasets/socialNetwork/scripts

# mediaMicroservices: TMDB datasets + write_movie_info.py (used by data-init-job
# and load-generator deployment to seed movie/cast/user data).
COPY ./mediaMicroservices/datasets /datasets/mediaMicroservices/datasets
COPY ./mediaMicroservices/scripts  /datasets/mediaMicroservices/scripts

# This image is content-only; it is never executed as a long-running process.
# The chart init container overrides this with `sh -c 'cp -r /datasets/. /shared/'`.
ENTRYPOINT ["/bin/true"]
