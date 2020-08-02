package client

import (
	"context"
	"time"

	goclient "github.com/micro/go-micro/v3/client"
	"github.com/micro/go-micro/v3/registry"
	"github.com/micro/micro/v3/service/client"
	"github.com/micro/micro/v3/service/errors"
	pb "github.com/micro/micro/v3/service/registry/proto"
	"github.com/micro/micro/v3/service/registry/util"
)

type srv struct {
	opts registry.Options
	// name of the registry
	name string
	// address
	address []string
	// client to call registry
	client pb.RegistryService
}

func (s *srv) callOpts() []goclient.CallOption {
	var opts []goclient.CallOption

	// set registry address
	if len(s.address) > 0 {
		opts = append(opts, goclient.WithAddress(s.address...))
	}

	// set timeout
	if s.opts.Timeout > time.Duration(0) {
		opts = append(opts, goclient.WithRequestTimeout(s.opts.Timeout))
	}

	return opts
}

func (s *srv) Init(opts ...registry.Option) error {
	for _, o := range opts {
		o(&s.opts)
	}

	if len(s.opts.Addrs) > 0 {
		s.address = s.opts.Addrs
	}

	return nil
}

func (s *srv) Options() registry.Options {
	return s.opts
}

func (s *srv) Register(srv *registry.Service, opts ...registry.RegisterOption) error {
	var options registry.RegisterOptions
	for _, o := range opts {
		o(&options)
	}
	if options.Context == nil {
		options.Context = context.TODO()
	}

	// encode srv into protobuf and pack TTL and domain into it
	pbSrv := util.ToProto(srv)
	pbSrv.Options.Ttl = int64(options.TTL.Seconds())
	pbSrv.Options.Domain = options.Domain

	// register the service
	_, err := s.client.Register(options.Context, pbSrv, s.callOpts()...)
	return err
}

func (s *srv) Deregister(srv *registry.Service, opts ...registry.DeregisterOption) error {
	var options registry.DeregisterOptions
	for _, o := range opts {
		o(&options)
	}
	if options.Context == nil {
		options.Context = context.TODO()
	}

	// encode srv into protobuf and pack domain into it
	pbSrv := util.ToProto(srv)
	pbSrv.Options.Domain = options.Domain

	// deregister the service
	_, err := s.client.Deregister(options.Context, pbSrv, s.callOpts()...)
	return err
}

func (s *srv) GetService(name string, opts ...registry.GetOption) ([]*registry.Service, error) {
	var options registry.GetOptions
	for _, o := range opts {
		o(&options)
	}
	if options.Context == nil {
		options.Context = context.TODO()
	}

	rsp, err := s.client.GetService(options.Context, &pb.GetRequest{
		Service: name, Options: &pb.Options{Domain: options.Domain},
	}, s.callOpts()...)

	if verr := errors.Parse(err); verr != nil && verr.Code == 404 {
		return nil, registry.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	services := make([]*registry.Service, 0, len(rsp.Services))
	for _, service := range rsp.Services {
		services = append(services, util.ToService(service))
	}
	return services, nil
}

func (s *srv) ListServices(opts ...registry.ListOption) ([]*registry.Service, error) {
	var options registry.ListOptions
	for _, o := range opts {
		o(&options)
	}
	if options.Context == nil {
		options.Context = context.TODO()
	}

	req := &pb.ListRequest{Options: &pb.Options{Domain: options.Domain}}
	rsp, err := s.client.ListServices(options.Context, req, s.callOpts()...)
	if err != nil {
		return nil, err
	}

	services := make([]*registry.Service, 0, len(rsp.Services))
	for _, service := range rsp.Services {
		services = append(services, util.ToService(service))
	}

	return services, nil
}

func (s *srv) Watch(opts ...registry.WatchOption) (registry.Watcher, error) {
	var options registry.WatchOptions
	for _, o := range opts {
		o(&options)
	}
	if options.Context == nil {
		options.Context = context.TODO()
	}

	stream, err := s.client.Watch(options.Context, &pb.WatchRequest{
		Service: options.Service, Options: &pb.Options{Domain: options.Domain},
	}, s.callOpts()...)

	if err != nil {
		return nil, err
	}

	return newWatcher(stream), nil
}

func (s *srv) String() string {
	return "service"
}

// NewRegistry returns a new registry service client
func NewRegistry(opts ...registry.Option) registry.Registry {
	var options registry.Options
	for _, o := range opts {
		o(&options)
	}

	// the registry address
	addrs := options.Addrs

	// don't default the address if a proxy is being used, as the
	// address will take precedent, circumventing the proxy.
	if len(addrs) == 0 && len(client.DefaultClient.Options().Proxy) == 0 {
		addrs = []string{"127.0.0.1:8000"}
	}

	name := "go.micro.registry"
	return &srv{
		opts:    options,
		name:    name,
		address: addrs,
		client:  pb.NewRegistryService(name),
	}
}
