# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.23
ARG DEBIAN_VERSION=bookworm
ARG NSJAIL_VERSION=3.4

# =========================================================
# Build nsjail from source + golangci-lint
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

# Build nsjail
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
# Tools image for tests and linting
# =========================================================
FROM builder AS tools

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libnl-route-3-200 \
    libprotobuf32 \
    python3 \
    g++ \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Copy nsjail and golangci-lint from nsjail-builder
COPY --from=nsjail-builder /usr/local/bin/nsjail /usr/local/bin/nsjail
COPY --from=nsjail-builder /usr/local/bin/golangci-lint /usr/local/bin/golangci-lint

WORKDIR /src

# =========================================================
# Runtime image (final production image)
# =========================================================
FROM debian:${DEBIAN_VERSION}-slim AS runtime

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libnl-route-3-200 \
    libprotobuf32 \
    python3 \
    g++ \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install golangci-lint in runtime too (optional but useful for consistency)
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
    sh -s -- -b /usr/local/bin v1.64.5

COPY --from=nsjail-builder /usr/local/bin/nsjail /usr/local/bin/nsjail
COPY --from=builder /out/goboxd /usr/local/bin/goboxd

WORKDIR /app

COPY languages.yaml .
COPY config/ /app/config/

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/goboxd"]