FROM golang:1.12.5-alpine as builder
LABEL maintainer="Yui Qi Tang <yqtang1222@gmail.com>"
ENV SRC=/src
COPY ./src $SRC
RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh
RUN cd $SRC && go build -o sample ./main.go

FROM alpine
LABEL maintainer="Yui Qi Tang <yqtang1222@gmail.com>"
WORKDIR /app
#       image       source       dest
COPY --from=builder /src/sample /app/
EXPOSE  8001
RUN cd /app
CMD ["./sample"]

# ENTRYPOINT ./app