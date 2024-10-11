FROM registry.cn-hangzhou.aliyuncs.com/117503445-mirror/sync:linux.amd64.docker.io.library.golang.1.23

RUN go env -w GOPROXY=https://goproxy.cn,direct

WORKDIR /workspace

ENTRYPOINT [ "bash" ]