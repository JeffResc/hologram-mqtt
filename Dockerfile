FROM --platform=$BUILDPLATFORM golang:1.26-alpine@sha256:70b46548e42db77e0966aaf3619fd068734dc6c77584d526b91126504fd95816 AS builder

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

FROM gcr.io/distroless/static@sha256:9197324ba51d9cd071af8505989365c006adf9d6d2067eada25aef00abbb5278

COPY --from=builder /hologram-mqtt /hologram-mqtt

ENTRYPOINT ["/hologram-mqtt"]
