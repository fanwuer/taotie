FROM golang:1.13-alpine AS go-build

WORKDIR /go/src/taotie

COPY core /go/src/taotie/core
COPY vendor /go/src/taotie/vendor
COPY main.go /go/src/taotie/main.go

RUN go build -ldflags "-s -w" -v -o taotie main.go

FROM alpine:3.10 AS prod

WORKDIR /root/

COPY --from=go-build /go/src/taotie/taotie /root/taotie_binary
RUN chmod 777 /root/taotie_binary
CMD /root/taotie_binary $RUN_OPTS