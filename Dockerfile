FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /pkgbadge .

FROM gcr.io/distroless/static-debian12
COPY --from=build /pkgbadge /pkgbadge
EXPOSE 8080
ENTRYPOINT ["/pkgbadge"]
