package main

import (
	"os"
	"path"

	"github.com/micro/micro/v3/service"
	plugin "github.com/micro/micro/v3/service/debug/collector/micro"
	log "github.com/micro/micro/v3/service/logger"
	"github.com/netdata/go-orchestrator"
	"github.com/netdata/go-orchestrator/cli"
	"github.com/netdata/go-orchestrator/pkg/multipath"
)

var (
	cd, _         = os.Getwd()
	netdataConfig = multipath.New(
		os.Getenv("NETDATA_USER_CONFIG_DIR"),
		os.Getenv("NETDATA_STOCK_CONFIG_DIR"),
		path.Join(cd, "/../../../../etc/netdata"),
		path.Join(cd, "/../../../../usr/lib/netdata/conf.d"),
	)
)

func main() {
	// New Service
	srv := service.New(
		service.Name("go.micro.debug.collector"),
		service.Version("latest"),
	)

	if len(os.Args) > 1 {
		os.Args = append(os.Args[:1], os.Args[2:]...)
	}

	// Initialise service
	srv.Init()

	go func() {
		log.Fatal(srv.Run())
	}()

	// register the new plugin
	plugin.New(srv.Client()).Register()

	netdata := orchestrator.New()
	netdata.Name = "micro.d"
	netdata.Option = &cli.Option{
		UpdateEvery: 1,
		Debug:       true,
		Module:      "all",
		ConfigDir:   netdataConfig,
		Version:     false,
	}
	netdata.ConfigPath = netdataConfig

	if !netdata.Setup() {
		log.Fatal("Netdata failed to Setup()")
	}

	netdata.Serve()
}
