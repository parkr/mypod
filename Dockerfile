FROM golang:1.23.5-alpine as builder
RUN set -ex \
  && wget -O /etc/mime.types 'https://svn.apache.org/viewvc/httpd/httpd/trunk/docs/conf/mime.types?revision=1901273&view=co'
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY cmd/ cmd/
COPY *.go ./
RUN set -ex \
  && CGO_ENABLED=0 go install ./... \
  && CGO_ENABLED=0 go test ./... \
  && ls /go/bin

FROM parkr/yt-dlp:2025.01.15
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
