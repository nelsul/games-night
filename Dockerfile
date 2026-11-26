# Builds the Svelte app and the Go API into one small image.
FROM node:22-alpine AS web
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY web/ ./
RUN npm run build

FROM golang:1.27-alpine AS api
WORKDIR /api
COPY api/go.mod api/go.sum ./
RUN go mod download
COPY api/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/games-night .

FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata && adduser -D -u 10001 app
ENV TZ=America/Sao_Paulo STATIC_DIR=/app/static ADDR=:8080
COPY --from=api /out/games-night /app/games-night
COPY --from=web /web/build /app/static
USER app
EXPOSE 8080
CMD ["/app/games-night"]
