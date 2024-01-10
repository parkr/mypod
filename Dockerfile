FROM golang:1.21.6-alpine as builder
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

FROM wernight/youtube-dl
COPY --from=builder /etc/mime.types /etc/mime.types
RUN set -ex \
  && youtube-dl --update \
  && apk add --no-cache attr wget
# Remove this once there's been a release which includes
# https://github.com/ytdl-org/youtube-dl/pull/31675
# Tracking bug: 
RUN set -ex \
  && rm /usr/local/bin/youtube-dl \
  && apk add --no-cache git python3 py3-pip \
  && python3 -m  pip install --verbose --upgrade --force-reinstall https://github.com/ytdl-org/youtube-dl/archive/refs/heads/master.tar.gz \
  && which youtube-dl
RUN set -ex \
  && wget https://github.com/wez/atomicparsley/releases/download/20210715.151551.e7ad03a/AtomicParsleyAlpine.zip \
  && unzip AtomicParsleyAlpine.zip \
  && mv AtomicParsley /usr/local/bin/AtomicParsley \
  && rm AtomicParsleyAlpine.zip
WORKDIR /storage
COPY --from=builder /go/bin/mypod /bin/mypod
ENTRYPOINT [ "/bin/mypod" ]
CMD [ "-h" ]
