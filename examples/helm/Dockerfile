FROM --platform=${BUILDPLATFORM} docker.io/library/golang:1.24 AS build
ARG TARGETPLATFORM
ARG BUILDPLATFORM

RUN apt-get update && apt-get install -y curl clang

RUN curl https://sh.rustup.rs -sSf | sh -s -- -y --default-toolchain stable
RUN . "$HOME/.cargo/env"

ENV PATH="/root/.cargo/bin:/usr/local/bin:$PATH"

WORKDIR /usr/src
COPY . .

RUN cargo build --release

FROM --platform=${BUILDPLATFORM} docker.io/library/golang:1.24
WORKDIR /apps
RUN mkdir /.config
COPY --from=build /usr/src/target/release/example-helm /apps
ENTRYPOINT ["/apps/example-helm"]