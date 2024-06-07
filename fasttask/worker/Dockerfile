ARG DEBIAN_RELEASE=bookworm

FROM rust:1.76-${DEBIAN_RELEASE} as builder
WORKDIR /usr/src/worker
COPY . .
RUN apt update && apt install -y python3-dev && cargo install --path .

FROM debian:${DEBIAN_RELEASE}-slim
COPY --from=builder /usr/local/cargo/bin/worker /usr/local/bin/worker
CMD ["worker"]
