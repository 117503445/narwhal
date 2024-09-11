FROM rust:1.62-bullseye AS builder
ARG PROFILE=release
ARG GIT_REVISION
ENV GIT_REVISION=$GIT_REVISION
COPY ./Docker/sources.list /etc/apt/sources.list
RUN apt-get update && apt-get install -y cmake clang

ENV RUSTUP_DIST_SERVER="https://rsproxy.cn"
ENV RUSTUP_UPDATE_ROOT="https://rsproxy.cn/rustup"

COPY ./Docker/config.toml /root/.cargo/config

ENTRYPOINT [ "sleep", "infinity" ]

WORKDIR /workspace