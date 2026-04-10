FROM node:20-alpine AS frontend
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.24-alpine AS backend
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/web/dist ./cmd/server/dist
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o /server ./cmd/server/

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend /server .
RUN mkdir -p /app/data

EXPOSE 8080
VOLUME ["/app/data"]

ENV DATABASE_PATH=/app/data/rss.db
ENV TZ=Asia/Shanghai

ENTRYPOINT ["/app/server"]
