FROM golang:alpine

WORKDIR /go/src/snyk-onboard
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

RUN mkdir /repos && chmod 700 /repos

CMD ["snyk-onboard"]
