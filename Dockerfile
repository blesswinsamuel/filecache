FROM --platform=$BUILDPLATFORM golang:1.18-alpine AS builder

WORKDIR /app

# RUN apk add --no-cache git

# RUN go get github.com/githubnemo/CompileDaemon

# COPY go.mod go.sum ./
COPY go.mod ./
RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o filecache .

FROM --platform=$BUILDPLATFORM alpine

# Copy our static executable.
COPY --from=builder /app/filecache /go/bin/filecache
# Run the hello binary.

CMD ["/go/bin/filecache"]
