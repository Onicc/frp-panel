# syntax=docker/dockerfile:1.7

FROM --platform=$BUILDPLATFORM node:24.10.0-alpine AS web
WORKDIR /src/www
RUN corepack enable && corepack prepare pnpm@10.17.1 --activate
COPY www/package.json www/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY www/index.html www/tsconfig.json www/vite.config.ts ./
COPY www/public ./public
COPY www/src ./src
RUN pnpm build

FROM --platform=$BUILDPLATFORM golang:1.27.1-alpine AS go-base
WORKDIR /src
RUN apk add --no-cache ca-certificates git tzdata
COPY go.mod go.sum ./
RUN go mod download

FROM go-base AS master-build
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=1970-01-01T00:00:00Z
COPY . .
COPY --from=web /src/cmd/frpp/out ./cmd/frpp/out
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath \
    -ldflags="-s -w -X github.com/Onicc/frp-panel/conf.gitVersion=${VERSION} -X github.com/Onicc/frp-panel/conf.gitCommit=${COMMIT} -X github.com/Onicc/frp-panel/conf.buildDate=${BUILD_DATE}" \
    -o /out/frp-panel ./cmd/frpp

FROM go-base AS agent-build
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=1970-01-01T00:00:00Z
COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath \
    -ldflags="-s -w -X github.com/Onicc/frp-panel/conf.gitVersion=${VERSION} -X github.com/Onicc/frp-panel/conf.gitCommit=${COMMIT} -X github.com/Onicc/frp-panel/conf.buildDate=${BUILD_DATE}" \
    -o /out/frp-panel-agent ./cmd/frp-panel-agent

FROM alpine:3.23 AS master
RUN apk add --no-cache ca-certificates tzdata && addgroup -S -g 10001 frp-panel && adduser -S -D -H -u 10001 -G frp-panel frp-panel && mkdir -p /data && chown frp-panel:frp-panel /data
COPY --from=master-build /out/frp-panel /usr/local/bin/frp-panel
USER 10001:10001
VOLUME ["/data"]
EXPOSE 9000 9001
ENV DB_TYPE=sqlite3 DB_DSN=/data/frp-panel.db?_pragma=journal_mode(WAL)
ENTRYPOINT ["/usr/local/bin/frp-panel"]
CMD ["master"]

FROM alpine:3.23 AS agent
RUN apk add --no-cache ca-certificates tzdata && addgroup -S -g 10001 frp-panel && adduser -S -D -H -u 10001 -G frp-panel frp-panel && mkdir -p /etc/frp-panel /var/lib/frp-panel && chown -R frp-panel:frp-panel /etc/frp-panel /var/lib/frp-panel
COPY --from=agent-build /out/frp-panel-agent /usr/local/bin/frp-panel-agent
USER 10001:10001
VOLUME ["/var/lib/frp-panel"]
ENTRYPOINT ["/usr/local/bin/frp-panel-agent"]
CMD ["agent", "run", "--config", "/etc/frp-panel/agent.yaml"]
