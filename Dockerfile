FROM golang:alpine

RUN mkdir -p /go/src/app
ADD . /go/src/app
WORKDIR /go/src/app
RUN apk add --no-cache git
RUN go get -u github.com/golang/dep/...
RUN go get -d -v ./...
RUN dep ensure
# ENV PORT 3001
# ENV GOOGLE_APPLICATION_CREDENTIALS="/go/src/app/service-account.json"
CMD ["go", "run", "main.go"]
