package client

import (
	"io"
	"sync"

	goclient "github.com/micro/go-micro/v3/client"
	"github.com/micro/go-micro/v3/runtime"
	pb "github.com/micro/micro/v3/proto/runtime"
	"github.com/micro/micro/v3/service/client"
	"github.com/micro/micro/v3/service/context"
)

type svc struct {
	sync.RWMutex
	options runtime.Options
	runtime pb.RuntimeService
}

// Init initializes runtime with given options
func (s *svc) Init(opts ...runtime.Option) error {
	s.Lock()
	defer s.Unlock()

	for _, o := range opts {
		o(&s.options)
	}

	s.runtime = pb.NewRuntimeService("runtime", client.DefaultClient)

	return nil
}

// Create registers a service in the runtime
func (s *svc) Create(svc *runtime.Service, opts ...runtime.CreateOption) error {
	var options runtime.CreateOptions
	for _, o := range opts {
		o(&options)
	}

	// set the default source from MICRO_RUNTIME_SOURCE
	if len(svc.Source) == 0 {
		svc.Source = s.options.Source
	}

	// runtime service create request
	req := &pb.CreateRequest{
		Service: &pb.Service{
			Name:     svc.Name,
			Version:  svc.Version,
			Source:   svc.Source,
			Metadata: svc.Metadata,
		},
		Options: &pb.CreateOptions{
			Command:    options.Command,
			Args:       options.Args,
			Env:        options.Env,
			Type:       options.Type,
			Image:      options.Image,
			Namespace:  options.Namespace,
			Secrets:    options.Secrets,
			Entrypoint: options.Entrypoint,
		},
	}

	if _, err := s.runtime.Create(context.DefaultContext, req, goclient.WithAuthToken()); err != nil {
		return err
	}

	return nil
}

func (s *svc) Logs(service *runtime.Service, opts ...runtime.LogsOption) (runtime.Logs, error) {
	var options runtime.LogsOptions
	for _, o := range opts {
		o(&options)
	}

	ls, err := s.runtime.Logs(context.DefaultContext, &pb.LogsRequest{
		Service: service.Name,
		Stream:  options.Stream,
		Count:   options.Count,
		Options: &pb.LogsOptions{
			Namespace: options.Namespace,
		},
	}, goclient.WithAuthToken())
	if err != nil {
		return nil, err
	}
	logStream := &serviceLogs{
		service: service.Name,
		stream:  make(chan runtime.Log),
		stop:    make(chan bool),
	}

	go func() {
		for {
			select {
			// @todo this never seems to return, investigate
			case <-ls.Context().Done():
				logStream.Stop()
			}
		}
	}()

	go func() {
		for {
			select {
			// @todo this never seems to return, investigate
			case <-ls.Context().Done():
				return
			case _, ok := <-logStream.stream:
				if !ok {
					return
				}
			default:
				record := pb.LogRecord{}
				err := ls.RecvMsg(&record)
				if err != nil {
					if err != io.EOF {
						logStream.err = err
					}
					logStream.Stop()
					return
				}
				logStream.stream <- runtime.Log{
					Message:  record.GetMessage(),
					Metadata: record.GetMetadata(),
				}
			}
		}
	}()
	return logStream, nil
}

type serviceLogs struct {
	service string
	stream  chan runtime.Log
	sync.Mutex
	stop chan bool
	err  error
}

func (l *serviceLogs) Error() error {
	return l.err
}

func (l *serviceLogs) Chan() chan runtime.Log {
	return l.stream
}

func (l *serviceLogs) Stop() error {
	l.Lock()
	defer l.Unlock()
	select {
	case <-l.stop:
		return nil
	default:
		close(l.stream)
		close(l.stop)
	}
	return nil
}

// Read returns the service with the given name from the runtime
func (s *svc) Read(opts ...runtime.ReadOption) ([]*runtime.Service, error) {
	var options runtime.ReadOptions
	for _, o := range opts {
		o(&options)
	}

	// runtime service create request
	req := &pb.ReadRequest{
		Options: &pb.ReadOptions{
			Service:   options.Service,
			Version:   options.Version,
			Type:      options.Type,
			Namespace: options.Namespace,
		},
	}

	resp, err := s.runtime.Read(context.DefaultContext, req, goclient.WithAuthToken())
	if err != nil {
		return nil, err
	}

	services := make([]*runtime.Service, 0, len(resp.Services))
	for _, service := range resp.Services {
		svc := &runtime.Service{
			Name:     service.Name,
			Version:  service.Version,
			Source:   service.Source,
			Metadata: service.Metadata,
			Status:   runtime.ServiceStatus(service.Status),
		}
		services = append(services, svc)
	}

	return services, nil
}

// Update updates the running service
func (s *svc) Update(svc *runtime.Service, opts ...runtime.UpdateOption) error {
	var options runtime.UpdateOptions
	for _, o := range opts {
		o(&options)
	}
	// runtime service create request
	req := &pb.UpdateRequest{
		Service: &pb.Service{
			Name:     svc.Name,
			Version:  svc.Version,
			Source:   svc.Source,
			Metadata: svc.Metadata,
		},
		Options: &pb.UpdateOptions{
			Namespace:  options.Namespace,
			Entrypoint: options.Entrypoint,
		},
	}

	if _, err := s.runtime.Update(context.DefaultContext, req, goclient.WithAuthToken()); err != nil {
		return err
	}

	return nil
}

// Delete stops and removes the service from the runtime
func (s *svc) Delete(svc *runtime.Service, opts ...runtime.DeleteOption) error {
	var options runtime.DeleteOptions
	for _, o := range opts {
		o(&options)
	}

	// runtime service dekete request
	req := &pb.DeleteRequest{
		Service: &pb.Service{
			Name:     svc.Name,
			Version:  svc.Version,
			Source:   svc.Source,
			Metadata: svc.Metadata,
		},
		Options: &pb.DeleteOptions{
			Namespace: options.Namespace,
		},
	}

	if _, err := s.runtime.Delete(context.DefaultContext, req, goclient.WithAuthToken()); err != nil {
		return err
	}

	return nil
}

func (s *svc) CreateNamespace(ns string) error {
	req := &pb.CreateNamespaceRequest{
		Namespace: ns,
	}
	if _, err := s.runtime.CreateNamespace(context.DefaultContext, req, goclient.WithAuthToken()); err != nil {
		return err
	}

	return nil
}

func (s *svc) DeleteNamespace(ns string) error {
	req := &pb.DeleteNamespaceRequest{
		Namespace: ns,
	}
	if _, err := s.runtime.DeleteNamespace(context.DefaultContext, req, goclient.WithAuthToken()); err != nil {
		return err
	}

	return nil
}

// Start starts the runtime
func (s *svc) Start() error {
	// NOTE: nothing to be done here
	return nil
}

// Stop stops the runtime
func (s *svc) Stop() error {
	// NOTE: nothing to be done here
	return nil
}

// Returns the runtime service implementation
func (s *svc) String() string {
	return "service"
}

// NewRuntime creates new service runtime and returns it
func NewRuntime(opts ...runtime.Option) runtime.Runtime {
	var options runtime.Options
	for _, o := range opts {
		o(&options)
	}

	return &svc{
		options: options,
		runtime: pb.NewRuntimeService("runtime", client.DefaultClient),
	}
}
