FROM registry.cn-hangzhou.aliyuncs.com/117503445-mirror/dev-golang

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && go install github.com/twitchtv/twirp/protoc-gen-twirp@latest
RUN pacman -Sy --noconfirm protobuf