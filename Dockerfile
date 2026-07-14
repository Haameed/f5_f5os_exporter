# syntax=docker/dockerfile:1

# ---- Build stage ----
FROM golang:1.25-alpine AS build

WORKDIR /src

# Cache dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .

ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_DATE

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w \
      -X main.version=${VERSION} \
      -X main.commit=${COMMIT} \
      -X main.buildDate=${BUILD_DATE}" \
    -o /bin/f5_f5os_exporter ./cmd/f5_f5os_exporter

# ---- Runtime stage ----
FROM gcr.io/distroless/static-debian12:nonroot

LABEL org.opencontainers.image.title="f5_f5os_exporter" \
      org.opencontainers.image.description="Prometheus exporter for F5 BIG-IP F5os devices" \
      org.opencontainers.image.source="https://github.com/Haameed/f5_f5os_exporter" \
      org.opencontainers.image.licenses="MIT"

COPY --from=build /bin/f5_f5os_exporter /bin/f5_f5os_exporter

EXPOSE 11001
USER nonroot:nonroot

ENTRYPOINT ["/bin/f5_f5os_exporter"]
CMD ["-config", "/etc/f5_f5os_exporter/config.yaml"]
