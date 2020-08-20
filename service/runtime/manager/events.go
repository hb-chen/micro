package manager

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	goauth "github.com/micro/go-micro/v3/auth"
	gorun "github.com/micro/go-micro/v3/runtime"
	gostore "github.com/micro/go-micro/v3/store"
	"github.com/micro/micro/v3/internal/namespace"
	"github.com/micro/micro/v3/service/client"
	"github.com/micro/micro/v3/service/logger"
	"github.com/micro/micro/v3/service/runtime"
	"github.com/micro/micro/v3/service/store"
)

var (
	// eventTTL is the duration events will perist in the store before expiring
	eventTTL = time.Minute * 10
	// eventPollFrequency is the max frequency the manager will check for new events in the store
	eventPollFrequency = time.Minute
)

const (
	// eventPrefix is prefixed to the key for event records
	eventPrefix = "event/"
	// eventProcessedPrefix is prefixed to the key for tracking event processing
	eventProcessedPrefix = "processed/"
)

// publishEvent will write the event to the global store and immediately process the event
func (m *manager) publishEvent(eType gorun.EventType, srv *gorun.Service, opts *gorun.CreateOptions) error {
	e := &gorun.Event{
		ID:      uuid.New().String(),
		Type:    eType,
		Service: srv,
		Options: opts,
	}

	bytes, err := json.Marshal(e)
	if err != nil {
		return err
	}

	record := &gostore.Record{
		Key:    eventPrefix + e.ID,
		Value:  bytes,
		Expiry: eventTTL,
	}

	if err := store.Write(record); err != nil {
		return err
	}

	go m.processEvent(record.Key)
	return nil
}

// watchEvents polls the store for events periodically and processes them if they have not already
// done so
func (m *manager) watchEvents() {
	ticker := time.NewTicker(eventPollFrequency)

	for {
		// get the keys of the events
		events, err := store.Read(eventPrefix, gostore.ReadPrefix())
		if err != nil {
			logger.Warn("Error listing events: %v", err)
			continue
		}

		// loop through every event
		for _, ev := range events {
			logger.Debugf("Process Event: %v", ev.Key)
			m.processEvent(ev.Key)
		}

		<-ticker.C
	}
}

// processEvent will take an event key, verify it hasn't been consumed and then execute it. We pass
// the key and not the ID since the global store and the memory store use the same key prefix so there
// is not point stripping and then re-prefixing.
func (m *manager) processEvent(key string) {
	// check to see if the event has been processed before
	if _, err := m.fileCache.Read(eventProcessedPrefix + key); err != gostore.ErrNotFound {
		return
	}

	// lookup the event
	recs, err := store.Read(key)
	if err != nil {
		logger.Warnf("Error finding event %v: %v", key, err)
		return
	}
	var ev *gorun.Event
	if err := json.Unmarshal(recs[0].Value, &ev); err != nil {
		logger.Warnf("Error unmarshaling event %v: %v", key, err)
	}

	// determine the namespace
	ns := namespace.DefaultNamespace
	if ev.Options != nil && len(ev.Options.Namespace) > 0 {
		ns = ev.Options.Namespace
	}

	// log the event
	logger.Infof("Processing %v event for service %v:%v in namespace %v", ev.Type, ev.Service.Name, ev.Service.Version, ns)

	// apply the event to the managed runtime
	switch ev.Type {
	case gorun.Delete:
		err = runtime.Delete(ev.Service, gorun.DeleteNamespace(ns))
	case gorun.Update:
		err = runtime.Update(ev.Service, gorun.UpdateNamespace(ns))
	case gorun.Create:
		// generate an auth account for the service to use
		var acc *goauth.Account
		acc, err = m.generateAccount(ev.Service, ns)
		if err != nil {
			return
		}

		// construct the options
		options := []gorun.CreateOption{
			gorun.CreateImage(ev.Options.Image),
			gorun.CreateType(ev.Options.Type),
			gorun.CreateNamespace(ns),
			gorun.WithArgs(ev.Options.Args...),
			gorun.WithCommand(ev.Options.Command...),
			gorun.WithEnv(m.runtimeEnv(ev.Service, ev.Options)),
		}

		// inject the credentials into the service if present
		if len(acc.ID) > 0 && len(acc.Secret) > 0 {
			options = append(options, gorun.WithSecret("MICRO_AUTH_ID", acc.ID))
			options = append(options, gorun.WithSecret("MICRO_AUTH_SECRET", acc.Secret))
		}

		// add the secrets
		for key, value := range ev.Options.Secrets {
			options = append(options, gorun.WithSecret(key, value))
		}

		// create the service
		err = runtime.Create(ev.Service, options...)
	}

	// if there was an error update the status in the cache
	if err != nil {
		logger.Warnf("Error processing %v event for service %v:%v in namespace %v: %v", ev.Type, ev.Service.Name, ev.Service.Version, ns, err)
		ev.Service.Metadata = map[string]string{"status": "error", "error": err.Error()}
		m.cacheStatus(ns, ev.Service)
	} else if ev.Type != gorun.Delete {
		m.cacheStatus(ns, ev.Service)
	}

	// write to the store indicating the event has been consumed. We double the ttl to safely know the
	// event will expire before this record
	m.fileCache.Write(&gostore.Record{Key: eventProcessedPrefix + key, Expiry: eventTTL * 2})

}

// runtimeEnv returns the environment variables which should  be used when creating a service.
func (m *manager) runtimeEnv(srv *gorun.Service, options *gorun.CreateOptions) []string {
	setEnv := func(p []string, env map[string]string) {
		for _, v := range p {
			parts := strings.Split(v, "=")
			if len(parts) <= 1 {
				continue
			}
			env[parts[0]] = strings.Join(parts[1:], "=")
		}
	}

	// overwrite any values
	env := map[string]string{
		// ensure a profile for the services isn't set, they
		// should use the default RPC clients
		"MICRO_PROFILE": "",
		// pass the service's name and version
		"MICRO_SERVICE_NAME":    nameFromService(srv.Name),
		"MICRO_SERVICE_VERSION": srv.Version,
		// set the proxy for the service to use (e.g. micro network)
		// using the proxy which has been configured for the runtime
		"MICRO_PROXY": client.DefaultClient.Options().Proxy,
	}

	// bind to port 8080, this is what the k8s tcp readiness check will use
	if runtime.DefaultRuntime.String() == "kubernetes" {
		env["MICRO_SERVICE_ADDRESS"] = ":8080"
	}

	// set the env vars provided
	setEnv(options.Env, env)

	// set the service namespace
	if len(options.Namespace) > 0 {
		env["MICRO_NAMESPACE"] = options.Namespace
	}

	// create a new env
	var vars []string
	for k, v := range env {
		vars = append(vars, k+"="+v)
	}

	// setup the runtime env
	return vars
}

// if a service is run from directory "/test/foo", the service
// will register as foo
func nameFromService(name string) string {
	comps := strings.Split(name, "/")
	if len(comps) == 0 {
		return ""
	}
	return comps[len(comps)-1]
}
