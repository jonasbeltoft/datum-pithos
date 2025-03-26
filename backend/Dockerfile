FROM golang:1.24.1-bookworm

# Important: required for go-sqlite3
ENV CGO_ENABLED=1
RUN apt-get update
RUN apt-get install --no-install-recommends --assume-yes \
	build-essential \
	libsqlite3-dev

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -v -o /usr/local/bin/app ./...

EXPOSE 8000

CMD ["app"]