FROM golang:1.23-alpine

WORKDIR /short_link

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .    

RUN go build -o short_link cmd/app/main.go
RUN go build -o short_link_worker cmd/worker/main.go

CMD ["./short_link", "./short_link_worker"]