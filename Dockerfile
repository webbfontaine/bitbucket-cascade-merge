FROM alpine:edge AS builder
LABEL maintainer="Samuel Contesse <samuel.contesse@morphean.com>"
RUN apk add --update-cache alpine-sdk go cmake heimdal-dev python3 && rm -rf /var/cache/apk/*
WORKDIR /build
RUN wget https://github.com/libgit2/libgit2/archive/refs/tags/v1.5.1.tar.gz -O libgit2.tar.gz
RUN tar -xvf libgit2.tar.gz
RUN cd libgit2-1.5.1 && mkdir build && cd build && cmake .. && make install

COPY . .
RUN go mod download

RUN go build -o main .

FROM alpine:edge
RUN apk add --update-cache rm -rf /var/cache/apk/*
COPY --from=builder /usr/local/lib/pkgconfig/libgit2.pc /usr/local/lib/pkgconfig/libgit2.pc
COPY --from=builder  /usr/local/lib/libgit2* /usr/local/lib/
COPY --from=builder /build/main /
ENTRYPOINT ["/main"]
