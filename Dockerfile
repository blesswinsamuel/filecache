FROM golang:1.14-alpine AS builder

WORKDIR /app

# RUN apk add --no-cache git

# RUN go get github.com/githubnemo/CompileDaemon

# COPY go.mod go.sum ./
COPY go.mod ./
RUN go mod download

COPY . .

RUN go install .

FROM alpine

# Copy our static executable.
COPY --from=builder /go/bin/filecache /go/bin/filecache
# Run the hello binary.

CMD ["/go/bin/filecache"]
