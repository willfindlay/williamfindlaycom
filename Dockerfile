FROM golang:1.25-alpine AS build

ARG COMMIT=unknown
ARG BUILD_TIME=unknown

RUN apk add --no-cache ca-certificates
RUN mkdir -p /data/content && chown 65534:65534 /data/content

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build \
    -ldflags="-s -w -X github.com/willfindlay/williamfindlaycom/internal/version.Commit=$COMMIT -X github.com/willfindlay/williamfindlaycom/internal/version.BuildTime=$BUILD_TIME" \
    -o /server ./cmd/server

FROM gcr.io/distroless/static:nonroot

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /server /server
COPY --from=build --chown=nonroot:nonroot /data /data

USER nonroot:nonroot
EXPOSE 8080

ENTRYPOINT ["/server"]
