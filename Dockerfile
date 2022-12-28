FROM alpine:latest

WORKDIR /work

ADD ./Hui-TxState /work/main

CMD ["./main"]

