module fc-worker

go 1.23.1

require (
	github.com/117503445/goutils v0.0.0-20241023113855-7542b5286948
	github.com/rs/zerolog v1.33.0
	google.golang.org/grpc v1.67.1
	google.golang.org/protobuf v1.34.2
)

require (
	github.com/twitchtv/twirp v8.1.3+incompatible // indirect
	golang.org/x/net v0.28.0 // indirect
	golang.org/x/text v0.17.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240814211410-ddb44dafa142 // indirect
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/sys v0.25.0 // indirect
	q v0.0.0-00010101000000-000000000000
)

replace q => /workspace/q
