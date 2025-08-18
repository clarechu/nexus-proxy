FROM golang:1.25.0-alpine3.21 AS build


ENV HOME=/home/cmdb

WORKDIR /home/cmdb

ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct

RUN  mkdir -p "/home/cmdb"

COPY . .

RUN go mod tidy && go build -o nexus3-proxy

FROM alpine AS runner

ENV TZ=Asia/Shanghai
ENV HOME=/home/cmdb
ENV PATH=$PATH:/home/cmdb

RUN mkdir -p $HOME && \
    apk --no-cache add ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

COPY --from=build /home/cmdb/nexus3-proxy $HOME/nexus-proxy

WORKDIR /home/cmdb

CMD ["nexus-proxy", "server"]


