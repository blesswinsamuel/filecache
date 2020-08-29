FROM golang:1.14-alpine

WORKDIR /app

RUN apk add --no-cache git

RUN go get github.com/githubnemo/CompileDaemon

COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["CompileDaemon", "-build=go install .", "-command=sso-fosite", "-color=true"]
