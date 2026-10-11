FROM node:24-bookworm-slim AS web
WORKDIR /app/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.25-bookworm AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd/
COPY internal/ ./internal/
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /server ./cmd/server

FROM debian:bookworm-slim
RUN groupadd --gid 10001 careerquest \
    && useradd --uid 10001 --gid careerquest --no-create-home careerquest \
    && mkdir -p /app/web/dist /data \
    && chown careerquest:careerquest /data
WORKDIR /app
COPY --from=backend /server /app/server
COPY --from=web /app/web/dist /app/web/dist
USER careerquest
ENV CQ_ADDR=0.0.0.0:8080 CQ_DATABASE=/data/career-quest.db CQ_WEB_DIR=/app/web/dist
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=4s --start-period=10s --retries=3 CMD ["/app/server", "healthcheck"]
ENTRYPOINT ["/app/server"]
