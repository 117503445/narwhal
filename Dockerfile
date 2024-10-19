# # Dictates whether the underlying node should be built as a dev or
# # production binary. If empty, then the development mode will be built.
# # If passed as BUILD_MODE=--release, then the production will be built.
# ARG BUILD_MODE="release"

# #################################################################
# # Stage 0
# #################################################################
# FROM rust:1.62-slim-bullseye as chef
# COPY Docker/sources.list /etc/apt/sources.list
# WORKDIR "$WORKDIR/narwhal"
# ARG BUILD_MODE

# # Install basic dependencies
# RUN DEBIAN_FRONTEND=noninteractive apt-get -qq update \
#     && DEBIAN_FRONTEND=noninteractive TZ=Etc/UTC apt-get -qq install -y --no-install-recommends \
#     tzdata \
#     git \
#     ca-certificates \
#     curl \
#     build-essential \
#     libssl-dev \
#     pkg-config \
#     clang \
#     cmake > /dev/null 2>&1 \
#     && rm -rf /var/lib/apt/lists/*
# # Install the fmt
# RUN rustup component add rustfmt
# RUN echo "Will build image with mode: ${BUILD_MODE}"

# #################################################################
# # Stage 1: Planning
# #################################################################
# FROM chef as planner
# ARG BUILD_MODE

# # Plan out the 3rd-party dependencies that need to be built.
# #
# # This is done by:
# #   1. Copy in Cargo.toml, Cargo.lock, and the workspace-hack crate
# #   2. Removing all workspace crates, other than the workpsace-hack
# #      crate, from the workspace Cargo.toml file.
# #   3. Update the lockfile in order to reflect the changes to the
# #      root Cargo.toml file.
# COPY Cargo.toml Cargo.lock ./
# COPY workspace-hack workspace-hack
# RUN sed -i 's/^members = .*/members = ["workspace-hack"]/g' Cargo.toml \
#     && cargo metadata -q >/dev/null

# #################################################################
# # Stage 2 : Caching
# #################################################################
# # Build and cache all dependencies.
# #
# # In a fresh layer, copy in the "plan" generated by the planner
# # and run `cargo build` in order to create a caching Docker layer
# # with all dependencies built.
# FROM chef AS builder 
# ARG BUILD_MODE
# ARG FEATURES="celo,benchmark"

# COPY --from=planner /narwhal/Cargo.toml Cargo.toml
# COPY --from=planner /narwhal/Cargo.lock Cargo.lock
# COPY --from=planner /narwhal/workspace-hack workspace-hack
# RUN cargo build --${BUILD_MODE} --all-features

# #################################################################
# # Stage 2.5 : Building
# #################################################################
# # Copy in the rest of the crates (and an unmodified Cargo.toml and Cargo.lock)
# # and build the application. At this point no dependencies should need to be
# # built as they were built and cached by the previous layer.

# # Copy all the files in the workdir excluding everything
# # from the .dockerignore
# COPY . .

# # Build the binary named "node"
# RUN cargo build --${BUILD_MODE} --features ${FEATURES} --bin node --bin benchmark_client


# #################################################################
# # Stage 3 : Production image
# #################################################################

# # Creat another layer so we can re-use caching
# FROM debian:bullseye-slim
# ARG BUILD_MODE
# COPY Docker/sources.list /etc/apt/sources.list
# WORKDIR "$WORKDIR/narwhal"

# # Copy the Narwhal node binary to bin folder
# COPY --from=builder narwhal/target/${BUILD_MODE}/node bin/

# # This is used for testing a cluster by generating load.
# # We use this in our k8s cluster deployed alongside the workers and validators.
# COPY --from=builder narwhal/target/${BUILD_MODE}/benchmark_client bin/

# # Copy the entry point file
# COPY Docker/entry.sh ./

# # Now add the entry point
# CMD ./entry.sh


FROM registry.cn-hangzhou.aliyuncs.com/117503445-mirror/sync:linux.amd64.docker.io.library.debian.bullseye-slim
COPY Docker/sources.list /etc/apt/sources.list
WORKDIR "/workspace"

# Copy the Narwhal node binary to bin folder
COPY ./docker-target/debug/node bin/

# This is used for testing a cluster by generating load.
# We use this in our k8s cluster deployed alongside the workers and validators.
COPY ./docker-target/debug/benchmark_client bin/

COPY ./q/q bin/

COPY ./q/client bin/

# Copy the entry point file
COPY Docker/entry.sh ./

# 设置为 UTC-8 时区
# RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && echo "Asia/Shanghai" > /etc/timezone

# Now add the entry point
CMD ["./entry.sh"]