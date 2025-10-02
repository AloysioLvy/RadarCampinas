# Stage 1: build
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Copia módulos e baixa dependências
COPY go.mod go.sum ./
RUN go mod download

# Copia todo o código-fonte
COPY . .

# Compila o binário estático
RUN CGO_ENABLED=0 \
    GOOS=linux \
    go build -a -installsuffix cgo -o server ./backend/cmd/server

# Stage 2: imagem final
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /root/

COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]