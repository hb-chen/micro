package server

import (
	"github.com/micro/cli/v2"
	"github.com/micro/go-micro/v3/auth"
	"github.com/micro/go-micro/v3/store"
	"github.com/micro/go-micro/v3/util/token"
	"github.com/micro/go-micro/v3/util/token/jwt"
	"github.com/micro/micro/v3/service"
	pb "github.com/micro/micro/v3/service/auth/proto"
	authHandler "github.com/micro/micro/v3/service/auth/server/auth"
	rulesHandler "github.com/micro/micro/v3/service/auth/server/rules"
	log "github.com/micro/micro/v3/service/logger"
	mustore "github.com/micro/micro/v3/service/store"
)

const (
	name    = "go.micro.auth"
	address = ":8010"
)

// Run the auth service
func Run(ctx *cli.Context) error {
	srv := service.New(
		service.Name(name),
		service.Address(address),
	)

	// setup the handlers
	ruleH := &rulesHandler.Rules{}
	authH := &authHandler.Auth{}

	// setup the auth handler to use JWTs
	pubKey := ctx.String("auth_public_key")
	privKey := ctx.String("auth_private_key")
	if len(pubKey) > 0 || len(privKey) > 0 {
		authH.TokenProvider = jwt.NewTokenProvider(
			token.WithPublicKey(pubKey),
			token.WithPrivateKey(privKey),
		)
	}

	// set the handlers store
	mustore.DefaultStore.Init(store.Table("auth"))
	authH.Init(auth.Store(mustore.DefaultStore))
	ruleH.Init(auth.Store(mustore.DefaultStore))

	// register handlers
	pb.RegisterAuthHandler(authH)
	pb.RegisterRulesHandler(ruleH)
	pb.RegisterAccountsHandler(authH)

	// run service
	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
	return nil
}
