FROM golang:1.14

ENV SRC_DIR /go/src/github.com/choobot/choo-pos-backend/app/
WORKDIR ${SRC_DIR}
COPY app/ ${SRC_DIR}
RUN go get -d -v ./...
RUN go install -v ./...

CMD ["/go/bin/app"]