FROM golang:1.12-alpine AS builder
RUN apk --update --no-cache add git && go get github.com/golang/dep/cmd/dep
WORKDIR /go/src/github.com/zachomedia/composerrepo/
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -vendor-only
COPY . ./
RUN go install ./...

FROM alpine:3.9
RUN apk --update --no-cache add ca-certificates
COPY --from=builder /go/bin/repo /usr/bin/repo
EXPOSE 8080
ENTRYPOINT [ "/usr/bin/repo" ]
