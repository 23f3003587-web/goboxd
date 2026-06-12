# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.23
ARG DEBIAN_VERSION=bookworm
ARG NSJAIL_VERSION=3.4

# =========================================================
# Build nsjail from source
# =========================================================
FROM debian:${DEBIAN_VERSION}-slim AS nsjail-builder

ARG NSJAIL_VERSION

RUN apt-get update && apt-get install -y --no-install-recommends \
    autoconf \
    bison \
    ca-certificates \
    curl \
    flex \
    g++ \
    gcc \
    git \
    libnl-route-3-dev \
    libprotobuf-dev \
    libtool \
    make \
    pkg-config \
    protobuf-compiler \
    && rm -rf /var/lib/apt/lists/*

RUN git clone --depth 1 --branch ${NSJAIL_VERSION} \
    https://github.com/google/nsjail.git /src/nsjail \
    && make -C /src/nsjail \
    && install -m 0755 /src/nsjail/nsjail /usr/local/bin/nsjail

# =========================================================
# Go builder stage
# =========================================================
FROM golang:${GO_VERSION}-${DEBIAN_VERSION} AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/goboxd \
    ./cmd/goboxd

# =========================================================
# Runtime image (final production image)
# =========================================================
FROM debian:${DEBIAN_VERSION}-slim AS runtime

LABEL org.opencontainers.image.title="goboxd" \
      org.opencontainers.image.description="Sandboxed code execution service" \
      org.opencontainers.image.source="https://github.com/thesouldev/goboxd"

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libnl-route-3-200 \
    libprotobuf32 \
    python3 \
    wget \
    g++ \
    gcc \
    default-jdk \
    nodejs \
    iverilog \
    lua5.3 \
    && rm -rf /var/lib/apt/lists/*

COPY --from=nsjail-builder /usr/local/bin/nsjail /usr/local/bin/nsjail
COPY --from=builder /out/goboxd /usr/local/bin/goboxd

WORKDIR /app
COPY config/ /app/config/

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=30s --retries=3 \
  CMD wget -q --spider http://127.0.0.1:8080/healthz || exit 1
ENTRYPOINT ["/usr/local/bin/goboxd"]

# =============================================
# Tools stage — for linting, testing, development
# =============================================
FROM builder AS tools

RUN go install github.com/rakyll/hey@v0.1.4
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.5 && \
    ln -s $(go env GOPATH)/bin/golangci-lint /usr/local/bin/golangci-lint

WORKDIR /src
