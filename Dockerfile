FROM golang:1.14

WORKDIR /go/src/github.com/choobot/choo-pos-backend/app/
COPY ./app/ ./
RUN go get -d -v ./...
RUN go install -v ./...

CMD ["/go/bin/app"]