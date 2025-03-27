FROM golang:1.24

# # Add Maintainer Info
LABEL maintainer="<>"

RUN mkdir /app
ADD . /app
WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./

RUN go build -o /dictionary_app

EXPOSE 8080

CMD ["./dictionary_app"]