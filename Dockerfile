FROM ubuntu:latest AS builder

RUN apt-get update && apt-get install -y wget tar

RUN wget https://go.dev/dl/go1.22.2.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go1.22.2.linux-amd64.tar.gz \
    && rm go1.22.2.linux-amd64.tar.gz

ENV PATH="/usr/local/go/bin:${PATH}"

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o server cmd/main.go

FROM ubuntu:latest

WORKDIR /root/

# Copy the compiled binary from the builder stage
COPY --from=builder /app/server .
# copy from a total different directory: configs/config.ini /root/configs/config.ini -> allow multiple config file for multiple node
COPY configs/config.ini /root/configs/config.ini

EXPOSE 9000/TCP
EXPOSE 9001/TCP

CMD ["./server"]