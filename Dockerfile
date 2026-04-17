FROM golang:1.24-alpine AS build
ARG VERSION=dev
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=${VERSION}" -o /pkgbadge .

FROM gcr.io/distroless/static-debian12

ARG VERSION=dev
ARG BUILD_DATE=unknown

LABEL org.opencontainers.image.source="https://github.com/Will-Luck/pkgbadge" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.title="pkgbadge" \
      org.opencontainers.image.description="Self-hosted badge endpoint for GHCR container packages"

COPY --from=build /pkgbadge /pkgbadge
EXPOSE 8080
ENTRYPOINT ["/pkgbadge"]
