# Etap 1: Budowanie z maksymalną kompresją
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Instalujemy UPX
RUN apk add --no-cache upx

# Optymalizacje dla minimalnego rozmiaru
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Kopiujemy aplikację
COPY app.go .

# Minimalny healthcheck (skompresowany healthcheck - oneline file)
RUN echo 'package main;import("net";"os");func main(){c,e:=net.Dial("tcp","localhost:8080");if e!=nil{os.Exit(1)};c.Close();os.Exit(0)}' > hc.go

# Kompilujemy główną aplikację
RUN go build -ldflags="-s -w -extldflags '-static'" \
    -gcflags="all=-l -B" \
    -trimpath \
    -o zadanie1 app.go && \
    upx --ultra-brute --best zadanie1

# Kompilujemy minimalny healthcheck
RUN go build -ldflags="-s -w -extldflags '-static'" \
    -gcflags="all=-l -B -N" \
    -trimpath \
    -o hc hc.go && \
    upx --ultra-brute --best hc

# Etap 2: Absolutnie minimalny runtime
FROM scratch

# Kopiujemy aplikację i mini healthcheck
COPY --from=builder /build/zadanie1 /zadanie1
COPY --from=builder /build/hc /hc

# Metadata
LABEL org.opencontainers.image.title="zadanie1" \
      org.opencontainers.image.authors="Jakub Nowosad"

EXPOSE 8080

# Healthcheck
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/hc"]

ENTRYPOINT ["/zadanie1"]