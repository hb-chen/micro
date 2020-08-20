module example

go 1.13

require (
	github.com/golang/protobuf v1.4.2
	github.com/micro/go-micro/v3 v3.0.0-beta.0.20200817090452-870a1ebc07bb
	github.com/micro/micro/v3 v3.0.0-beta.0.20200817234352-e8d00c2dea0d
	google.golang.org/grpc v1.27.0
	google.golang.org/protobuf v1.25.0
)

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0

replace github.com/micro/micro/v3 => ../../..
