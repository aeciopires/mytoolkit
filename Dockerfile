# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

WORKDIR /build

COPY app/go.mod app/go.sum ./
RUN go mod download

COPY app/ ./

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w -X github.com/aeciopires/mytoolkit/internal/version.Version=$VERSION" \
    -o /out/mytoolkit ./cmd/mytoolkit

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/mytoolkit /mytoolkit

USER nonroot:nonroot
EXPOSE 8080

ENTRYPOINT ["/mytoolkit"]
CMD ["serve"]
