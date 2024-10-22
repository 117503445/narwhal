module fc-worker

go 1.23.1

require (
	github.com/117503445/goutils v0.0.0-20241022113537-38bd5e335a7e
	github.com/rs/zerolog v1.33.0
	google.golang.org/protobuf v1.34.2
)

require github.com/twitchtv/twirp v8.1.3+incompatible // indirect

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/sys v0.25.0 // indirect
	q v0.0.0-00010101000000-000000000000
)

replace q => /workspace/q
