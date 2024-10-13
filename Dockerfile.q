FROM 117503445/dev-golang

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
RUN pacman -S --noconfirm protobuf