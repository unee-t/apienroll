FROM golang:alpine

RUN apk --no-cache add git

ADD main.go /go/src/app/main.go
WORKDIR /go/src/app
RUN go get -d -v
RUN go install -v

FROM alpine:latest
COPY --from=0 /go/bin/app /go/bin/app

ARG COMMIT
ENV COMMIT ${COMMIT}

ENV PORT 9000
CMD ["/go/bin/app"]
