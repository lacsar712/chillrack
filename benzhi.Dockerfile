FROM golang:1.22

ENV PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
ENV GOTOOLCHAIN=local
ENV CGO_ENABLED=0

WORKDIR /app

COPY go.mod ./
COPY . .

RUN go build ./...

CMD ["bash"]
