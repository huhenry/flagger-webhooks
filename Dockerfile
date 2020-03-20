FROM alpine:3.10


LABEL maintainer='Henry Hu <huhenry.bj0@hotmail.com>'

ENV TZ=Asia/Shanghai
RUN apk --no-cache add curl && mkdir -p /grife

WORKDIR /grife

CMD ["/grife/grife"]

ENTRYPOINT ["/bin/sh", "-c", "/grife/grife"]

ADD bin/grife /grife/grife
