FROM amd64/alpine:latest

WORKDIR /work

ADD ./bin/linux-amd64-hui-collect /work/main

CMD ["./main"]

