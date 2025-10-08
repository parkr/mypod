FROM golang:1.25.2-alpine as builder
RUN set -ex \
  && wget -O /etc/mime.types 'https://raw.githubusercontent.com/apache/httpd/refs/heads/trunk/docs/conf/mime.types'
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY cmd/ cmd/
COPY *.go ./
RUN set -ex \
  && CGO_ENABLED=0 go install ./... \
  && CGO_ENABLED=0 go test ./... \
  && ls /go/bin

FROM parkr/yt-dlp:2025.09.26
COPY --from=builder /etc/mime.types /etc/mime.types
RUN which yt-dlp && yt-dlp --version
RUN set -ex \
  && wget https://github.com/wez/atomicparsley/releases/download/20210715.151551.e7ad03a/AtomicParsleyAlpine.zip \
  && unzip AtomicParsleyAlpine.zip \
  && mv AtomicParsley /usr/local/bin/AtomicParsley \
  && rm AtomicParsleyAlpine.zip
WORKDIR /storage
COPY --from=builder /go/bin/mypod /bin/mypod
ENTRYPOINT [ "/bin/mypod" ]
CMD [ "-h" ]
