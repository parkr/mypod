FROM golang:1.18.3-alpine as builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY cmd/ cmd/
COPY *.go ./
RUN set -ex \
  && CGO_ENABLED=0 go install ./... \
  && CGO_ENABLED=0 go test ./... \
  && ls /go/bin

FROM wernight/youtube-dl
RUN set -ex \
  && youtube-dl --update \
  && apk add --no-cache attr wget
RUN set -ex \
  && wget https://github.com/wez/atomicparsley/releases/download/20210715.151551.e7ad03a/AtomicParsleyAlpine.zip \
  && unzip AtomicParsleyAlpine.zip \
  && mv AtomicParsley /usr/local/bin/AtomicParsley \
  && rm AtomicParsleyAlpine.zip
RUN set -ex \
  && wget -O /etc/mime.types https://raw.githubusercontent.com/nginx/nginx/release-1.23.0/conf/mime.types
WORKDIR /storage
COPY --from=builder /go/bin/mypod /bin/mypod
ENTRYPOINT [ "/bin/mypod" ]
CMD [ "-h" ]
