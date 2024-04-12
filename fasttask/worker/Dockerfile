FROM rust:1.76-bookworm as builder
WORKDIR /usr/src/worker
COPY . .
RUN apt update && apt install -y python3-dev && cargo install --path .

FROM debian:bookworm-slim
COPY --from=builder /usr/local/cargo/bin/worker /usr/local/bin/worker
CMD ["worker"]
