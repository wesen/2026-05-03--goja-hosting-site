FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go test ./...
RUN CGO_ENABLED=1 go build -o /out/goja-site ./cmd/goja-site

FROM debian:bookworm-slim
RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates \
 && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /out/goja-site /usr/local/bin/goja-site
COPY sites /app/sites
COPY deploy/sites.yaml /etc/goja-site/sites.yaml
RUN useradd -r -u 10001 -g root goja-site \
 && mkdir -p /data/sites \
 && chown -R 10001:0 /data
USER 10001
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/goja-site"]
CMD ["serve-multi", "--config", "/etc/goja-site/sites.yaml"]
