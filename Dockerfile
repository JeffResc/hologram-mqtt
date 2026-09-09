FROM --platform=$BUILDPLATFORM golang:1.27-alpine@sha256:cf6fca6641884b8433441b2b0652976f975e1d0fdd26d177eaaf8596087f3125 AS builder

ARG TARGETARCH
ARG TARGETOS

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o /hologram-mqtt \
    ./cmd/hologram-mqtt

FROM gcr.io/distroless/static@sha256:f2ea2709ac8db56323cbd7d014277f32cb572d9ea124b0076f7aafe5980678fe

COPY --from=builder /hologram-mqtt /hologram-mqtt

ENTRYPOINT ["/hologram-mqtt"]
