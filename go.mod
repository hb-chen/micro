module github.com/micro/micro/v3

go 1.13

require (
	github.com/boltdb/bolt v1.3.1
	github.com/chzyer/logex v1.1.10 // indirect
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e
	github.com/chzyer/test v0.0.0-20180213035817-a1ea475d72b1 // indirect
	github.com/cloudflare/cloudflare-go v0.10.9 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/dustin/go-humanize v1.0.0
	github.com/fsnotify/fsnotify v1.4.7
	github.com/go-acme/lego/v3 v3.4.0
	github.com/golang/protobuf v1.4.2
	github.com/google/uuid v1.1.1
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.3
	github.com/hashicorp/go-version v1.2.1
	github.com/juju/fslock v0.0.0-20160525022230-4d5c94c67b4b
	github.com/micro/cli/v2 v2.1.2
	github.com/micro/go-micro/v3 v3.0.0-beta.0.20200820102804-329bc2f2655b
	github.com/olekukonko/tablewriter v0.0.4
	github.com/pkg/errors v0.9.1
	github.com/serenize/snaker v0.0.0-20171204205717-a683aaf2d516
	github.com/stretchr/testify v1.5.1
	github.com/xlab/treeprint v0.0.0-20181112141820-a009c3971eca
	golang.org/x/crypto v0.0.0-20200709230013-948cd5f35899
	golang.org/x/net v0.0.0-20200520182314-0ba52f642ac2
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013
	google.golang.org/grpc v1.27.0
)

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0
