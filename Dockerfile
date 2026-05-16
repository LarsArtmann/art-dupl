FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -tags netgo \
    -ldflags="-s -w -X github.com/LarsArtmann/art-dupl/cmd.Version=${VERSION} -X github.com/LarsArtmann/art-dupl/cmd.Commit=${COMMIT} -X github.com/LarsArtmann/art-dupl/cmd.Date=${BUILD_DATE}" \
    -trimpath \
    -o art-dupl \
    ./cmd/art-dupl

FROM gcr.io/distroless/static-debian13:nonroot

COPY --from=builder /build/art-dupl /art-dupl

USER 65532:65532

ENTRYPOINT ["/art-dupl"]
