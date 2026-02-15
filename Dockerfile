FROM golang:1.25-alpine AS build

RUN apk add --no-cache ca-certificates
RUN mkdir -p /data/content && chown 65534:65534 /data/content

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /server ./cmd/server

FROM gcr.io/distroless/static:nonroot

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /server /server
COPY --from=build --chown=nonroot:nonroot /data /data

USER nonroot:nonroot
EXPOSE 8080

ENTRYPOINT ["/server"]
