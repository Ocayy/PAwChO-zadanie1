# Sprawozdanie z zadania 1 - Programowanie Aplikacji w Chmurze Obliczeniowej

Autor: Jakub Nowosad

## Część Obowiązkowa

### 1. Aplikacja (max. 30%)

Aplikacja została napisana w języku Go i realizuje wszystkie wymagane funkcjonalności.

#### a. Logowanie informacji przy uruchomieniu

Po uruchomieniu kontenera, aplikacja zapisuje w logach:
- Datę uruchomienia
- Imię i nazwisko autora programu
- Port TCP, na którym aplikacja nasłuchuje

```go
// Fragment kodu odpowiedzialny za logowanie informacji startowych
func main() {
    if envPort := os.Getenv("PORT"); envPort != "" {
        port = envPort
    }

    fmt.Println("=== APLIKACJA URUCHOMIONA ===")
    fmt.Printf("Data uruchomienia: %s\n", time.Now().Format("2006-01-02 15:04:05"))
    fmt.Printf("Autor: %s\n", author)
    fmt.Printf("Port: %s\n", port)
    fmt.Println("============================")

    // ... reszta kodu
}
```

#### b. Funkcjonalność pogodowa

Aplikacja umożliwia:
- Wybór miasta z predefiniowanej listy (Warszawa, Kraków, Londyn, Paryż, Berlin)
- Zatwierdzenie wyboru poprzez przycisk
- Wyświetlenie aktualnej pogody w wybranej lokalizacji

Zakres wyświetlanych informacji pogodowych:
- Temperatura (°C)
- Wilgotność (%)
- Prędkość wiatru (km/h)

```go
// Struktura danych dla miast
var cities = map[string]City{
    "warszawa": {"Warszawa", 52.2297, 21.0122},
    "krakow":   {"Kraków", 50.0647, 19.9450},
    "londyn":   {"Londyn", 51.5074, -0.1278},
    "paryz":    {"Paryż", 48.8566, 2.3522},
    "berlin":   {"Berlin", 52.5200, 13.4050},
}

// Funkcja pobierająca dane pogodowe
func handleWeather(w http.ResponseWriter, r *http.Request) {
    // ... kod pobierający dane z API Open-Meteo
    url := fmt.Sprintf("http://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f&current_weather=true&hourly=relative_humidity_2m", city.Lat, city.Lon)
    // ... reszta implementacji
}
```

### 2. Dockerfile (max. 50%)

Plik Dockerfile został zoptymalizowany pod kątem minimalnego rozmiaru obrazu przy zachowaniu wszystkich funkcjonalności.

```dockerfile
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
```

#### Optymalizacje zastosowane w Dockerfile:
1. **Wieloetapowe budowanie** - używamy golang:alpine do budowania i scratch do runtime
2. **Warstwa scratch** - minimalizuje rozmiar końcowego obrazu do absolutnego minimum
3. **Optymalizacja cache** - kolejność komend optymalna dla cache'owania warstw
4. **Kompresja UPX** - ekstremalna kompresja pliku wykonywalnego
5. **Metadata OCI** - zgodność ze standardem OCI

### 3. Polecenia (max. 20%)

#### a. Zbudowanie obrazu kontenera
```bash
docker build -t zadanie1:v1 .
```

#### b. Uruchomienie kontenera
```bash
docker run -d -p 8080:8080 --name zadanie1-container zadanie1:v1
```

#### c. Uzyskanie informacji z logów
```bash
docker logs zadanie1-container
```

Przykładowy output z logów:
```
=== APLIKACJA URUCHOMIONA ===
Data uruchomienia: 2025-05-04 20:46:19
Autor: Jakub Nowosad
Port: 8080
============================
```

#### d. Sprawdzenie ilości warstw i rozmiaru obrazu

Sprawdzenie rozmiaru:
```bash
docker images zadanie1:v1
```

Output:
```
REPOSITORY   TAG       IMAGE ID       CREATED          SIZE
zadanie1     v1        f135954ea093   17 minutes ago   4.31MB
```

Sprawdzenie liczby warstw:
```bash
docker history zadanie1:v1
```

Output pokazuje minimalistyczną strukturę warstw dzięki użyciu scratch:
```
IMAGE          CREATED          CREATED BY                                      SIZE      COMMENT
f135954ea093   17 minutes ago   ENTRYPOINT ["/zadanie1"]                        0B        buildkit.dockerfile.v0
<missing>      17 minutes ago   HEALTHCHECK &{["CMD" "/hc"] "30s" "10s" "5s"…   0B        buildkit.dockerfile.v0
<missing>      17 minutes ago   EXPOSE map[8080/tcp:{}]                         0B        buildkit.dockerfile.v0
<missing>      17 minutes ago   LABEL org.opencontainers.image.title=zadanie…   0B        buildkit.dockerfile.v0
<missing>      17 minutes ago   COPY /build/hc /hc # buildkit                   573kB     buildkit.dockerfile.v0
<missing>      17 minutes ago   COPY /build/zadanie1 /zadanie1 # buildkit       1.59MB    buildkit.dockerfile.v0
```

Aplikacja jest dostępna pod adresem http://localhost:8080 i prezentuje:
1. Prosty interfejs z rozwijaną listą miast
2. Przycisk "Sprawdź pogodę"
3. Po kliknięciu wyświetla się panel z danymi pogodowymi

### Podsumowanie

Aplikacja została zrealizowana zgodnie z wymaganiami zadania:
- ✅ Loguje informacje przy starcie (data, autor, port)
- ✅ Umożliwia wybór miasta i sprawdzenie pogody
- ✅ Używa wieloetapowego budowania w Dockerfile
- ✅ Jest zoptymalizowana pod kątem rozmiaru (4.31MB), natomiast bez healthcheck (3.17MB)❗
- ✅ Zawiera metadata zgodne z OCI

Końcowy rozmiar obrazu wyniósł **4.31MB**, co jest bardzo dobrym wynikiem dla aplikacji Go z pełną funkcjonalnością.


### Linki

 [ **[GitHub](https://github.com/Ocayy/PAwChO-zadanie1)** ]  |  [ **[DockerHub](https://hub.docker.com/repository/docker/jaco13/politechnika-lubelska/tags/zadanie1/sha256-ba82354d2c8c3eaf0e6998ca24fe926b88a48930579b0a36880cebf0349dac2e)** ]