# Estagio 1: Compilacao
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /node ./cmd/node/main.go

# Estagio 2: Imagem minima (~5MB)
FROM alpine:3.19
COPY --from=builder /node /node
CMD ["/node"]
