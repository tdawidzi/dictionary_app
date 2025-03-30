FROM golang:1.24

# Ustawiamy maintainera
LABEL maintainer="<>"

# Ustawiamy katalog roboczy
WORKDIR /app

# Kopiujemy pliki modułu i pobieramy zależności
COPY go.mod go.sum ./
RUN go mod download

# Kopiujemy cały kod aplikacji
COPY . .

RUN go mod tidy

# Kompilujemy aplikację do katalogu /app
RUN go build -o dictionary_app

# Otwieramy port aplikacji
EXPOSE 8080

# Uruchamiamy skompilowaną aplikację
CMD ["/app/dictionary_app"]