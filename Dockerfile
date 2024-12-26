# Build aşaması
FROM golang:1.23-bullseye AS builder

# Gerekli build araçlarını yükle
RUN apt-get update && apt-get install -y git

# Çalışma dizinini ayarla
WORKDIR /app

# Go modüllerini kopyala ve indir
COPY go.mod go.sum ./
RUN go mod download

# Kaynak kodları kopyala
COPY . .

# Uygulamayı derle
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o sdc .

# Çalışma aşaması
FROM debian:bullseye-slim

# SSL sertifikaları için gerekli paketleri yükle
RUN apt-get update && apt-get install -y ca-certificates

WORKDIR /root/

# Builder aşamasından derlenmiş uygulamayı kopyala
COPY --from=builder /app/sdc .

# Uygulamayı çalıştır
CMD ["./sdc"] 