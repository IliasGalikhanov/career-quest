FROM node:24-bookworm-slim@sha256:d6aa754f16b3197301076f047b5def2f02ea1dbbc2ca920407d46d7ec7f87b20 AS web
WORKDIR /app/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.25-bookworm@sha256:3b4a11519ad929d1e1d261a12cff056f0c85b735253d7d861346b9c6f8b36437 AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd/
COPY internal/ ./internal/
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /server ./cmd/server

FROM debian:bookworm-slim@sha256:7c7b2c966bc9ee8cedfeef67e0e279108992c77681fa595db4a9d65c06ccc587
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
