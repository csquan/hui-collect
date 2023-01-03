FROM alpine:latest

WORKDIR /work

ADD ./HuiCollect /work/main

CMD ["./main"]

