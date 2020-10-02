package manager

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	gorun "github.com/micro/go-micro/v3/runtime"
	"github.com/micro/go-micro/v3/runtime/local/source/git"
	gostore "github.com/micro/go-micro/v3/store"
	"github.com/micro/micro/v3/service/auth"
	"github.com/micro/micro/v3/service/client"
	"github.com/micro/micro/v3/service/logger"
	"github.com/micro/micro/v3/service/runtime"
	"github.com/micro/micro/v3/service/runtime/builder"
	"github.com/micro/micro/v3/service/runtime/builder/golang"
	"github.com/micro/micro/v3/service/runtime/util/tar"
	"github.com/micro/micro/v3/service/store"
)

func (m *manager) buildAndRun(srv *service) {
	if err := m.build(srv); err != nil {
		return
	}

	srv.Status = runtime.Starting
	m.writeService(srv)

	if err := m.createServiceInRuntime(srv); err != nil {
		srv.Status = runtime.Error
		srv.Error = fmt.Sprintf("Error creating service: %v", err)
		m.writeService(srv)
	}
}

func (m *manager) buildAndUpdate(srv *service) {
	if err := m.build(srv); err != nil {
		return
	}

	srv.Status = runtime.Starting
	m.writeService(srv)

	if err := m.updateServiceInRuntime(srv); err != nil {
		srv.Status = runtime.Error
		srv.Error = fmt.Sprintf("Error updating service: %v", err)
		m.writeService(srv)
	}
}

func (m *manager) build(srv *service) error {
	// set the service status to building
	srv.Status = runtime.Building
	m.writeService(srv)

	// handleError will set the error status on the service
	handleError := func(err error, msg string) {
		srv.Status = runtime.Error
		srv.Error = fmt.Sprintf("%v: %v", msg, err)
		m.writeService(srv)
	}

	// load the source
	var source io.Reader
	var err error
	if strings.HasPrefix(srv.Service.Source, "source://") {
		// if the source was uploaded to the blob store, it'll have source:// as the prefix
		nsOpt := gostore.BlobNamespace(srv.Options.Namespace)
		source, err = store.DefaultBlobStore.Read(srv.Service.Source, nsOpt)
	} else {
		// the source will otherwise be a git remote, we'll clone it and then tar archive the result
		gitSrc, err := git.ParseSource(srv.Service.Source)
		if err != nil {
			handleError(err, "Error parsing git source")
			return err
		}
		if len(srv.Options.Entrypoint) == 0 {
			srv.Options.Entrypoint = gitSrc.Folder
		}

		// checkout the source
		gitSrc.Ref = srv.Service.Version
		dir, err := git.CheckoutSource(gitSrc, srv.Options.Secrets)
		if err != nil {
			handleError(err, "Error fetching git source")
			return err
		}

		// archive the source so it can be passed to the builder
		source, err = tar.Archive(dir)
	}
	if err != nil {
		handleError(err, "Error loading source")
		return err
	}

	// if we're building the build service, override the default builder implementation to prevent
	// a circular dependancy
	bldr := builder.DefaultBuilder
	if srv.Service.Source == "github.com/m3o/services/build" {
		bldr, _ = golang.NewBuilder()
	}

	// build the source
	build, err := bldr.Build(source,
		builder.Archive("tar"),
		builder.Entrypoint(srv.Options.Entrypoint),
	)
	if err != nil {
		handleError(err, "Error building service")
		return err
	}

	// for the kubernetes runtime, the container needs to pull the source (it's not got access to the
	// local filesystem like the local runtime does). hence we'll upload the source to the blob store
	// which the cell (container) can then pull via the Runtime.Build.Read RPC.
	if m.Runtime.String() != "local" {
		nsOpt := gostore.BlobNamespace(srv.Options.Namespace)
		key := fmt.Sprintf("build://%v:%v", srv.Service.Name, srv.Service.Version)
		if err := store.DefaultBlobStore.Write(key, build, nsOpt); err != nil {
			handleError(err, "Error uploading build")
			return err
		}
	}

	return nil
}

func (m *manager) updateServiceInRuntime(srv *service) error {
	// construct the options
	options := []gorun.UpdateOption{
		gorun.UpdateEntrypoint(srv.Options.Entrypoint),
		gorun.UpdateNamespace(srv.Options.Namespace),
	}

	// add the secrets
	for key, value := range srv.Options.Secrets {
		options = append(options, gorun.UpdateSecret(key, value))
	}

	// update the service
	return m.Runtime.Update(srv.Service, options...)
}

// createServiceInRuntime will add all the required env vars and secrets and then create the service
func (m *manager) createServiceInRuntime(srv *service) error {
	// generate an auth account for the service to use
	acc, err := m.generateAccount(srv)
	if err != nil {
		return err
	}

	// construct the options
	options := []gorun.CreateOption{
		gorun.CreateEntrypoint(srv.Options.Entrypoint),
		gorun.CreateImage(srv.Options.Image),
		gorun.CreateType(srv.Options.Type),
		gorun.CreateNamespace(srv.Options.Namespace),
		gorun.WithArgs(srv.Options.Args...),
		gorun.WithCommand(srv.Options.Command...),
		gorun.WithEnv(m.runtimeEnv(srv.Service, srv.Options)),
	}

	// add the secrets
	for key, value := range srv.Options.Secrets {
		options = append(options, gorun.WithSecret(key, value))
	}

	// inject the credentials into the service if present
	if len(acc.ID) > 0 && len(acc.Secret) > 0 {
		options = append(options, gorun.WithSecret("MICRO_AUTH_ID", acc.ID))
		options = append(options, gorun.WithSecret("MICRO_AUTH_SECRET", acc.Secret))
	}

	// create the service
	return m.Runtime.Create(srv.Service, options...)
}

// checkoutSource will take a service and download the source into a temp directory. This source
// could be a git remote or a reference to some source in the blob store (used for running local
// code on the server).
func (m *manager) checkoutSource(srv *service) (string, error) {
	if strings.HasPrefix(srv.Service.Source, "source://") {
		return m.checkoutBlobSource(srv)
	} else {
		return m.checkoutGitSource(srv)
	}
}

// checkoutBlobSource will checkout source from the blob store using the key in the service's source
// attribute. It will then unarchive the source into a temp directory and return the location of
// said directory.
func (m *manager) checkoutBlobSource(srv *service) (string, error) {
	nsOpt := gostore.BlobNamespace(srv.Options.Namespace)
	source, err := store.DefaultBlobStore.Read(srv.Service.Source, nsOpt)
	if err != nil {
		return "", err
	}

	dir, err := ioutil.TempDir(os.TempDir(), "blob-*")
	if err != nil {
		return "", err
	}

	if err := tar.Unarchive(source, dir); err != nil {
		return "", err
	}

	return dir, nil
}

// checkoutGitSource will download source from a git remote into a temp dir and then return the
// location of that temp directory
func (m *manager) checkoutGitSource(srv *service) (string, error) {
	gitSrc, err := git.ParseSource(srv.Service.Source)
	if err != nil {
		return "", err
	}
	gitSrc.Ref = srv.Service.Version

	dir, err := git.CheckoutSource(gitSrc, srv.Options.Secrets)
	if err != nil {
		return "", err
	}

	// the dir will contain the entire repo, however the use could've specified a subfolder within
	// that repo. this is the case for mono-repos
	if len(srv.Options.Entrypoint) == 0 {
		srv.Options.Entrypoint = gitSrc.Folder
	}

	return dir, nil
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
		"MICRO_SERVICE_NAME":    srv.Name,
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

func (m *manager) generateAccount(srv *service) (*auth.Account, error) {
	accName := srv.Service.Name + "-" + srv.Service.Version

	opts := []auth.GenerateOption{
		auth.WithIssuer(srv.Options.Namespace),
		auth.WithScopes("service"),
		auth.WithType("service"),
	}

	acc, err := auth.Generate(accName, opts...)
	if err != nil {
		if logger.V(logger.WarnLevel, logger.DefaultLogger) {
			logger.Warnf("Error generating account %v: %v", accName, err)
		}
		return nil, err
	}
	if logger.V(logger.DebugLevel, logger.DefaultLogger) {
		logger.Debugf("Generated auth account: %v, secret: [len: %v]", acc.ID, len(acc.Secret))
	}

	return acc, nil
}

// cleanupBlobStore deletes the source code and build from the blob store once the service finishes
// running.
func (m *manager) cleanupBlobStore(srv *service) {
	// delete the raw source code
	opt := gostore.BlobNamespace(srv.Options.Namespace)
	srcKey := fmt.Sprintf("source://%v:%v", srv.Service.Name, srv.Service.Version)
	if err := store.DefaultBlobStore.Delete(srcKey, opt); err != nil && err != store.ErrNotFound {
		logger.Warnf("Error deleting source %v: %v", srcKey, err)
	}

	// if there is no builder enabled, there won't be any build to delete
	if builder.DefaultBuilder == nil {
		return
	}

	// delete the binary
	buildKey := fmt.Sprintf("build://%v:%v", srv.Service.Name, srv.Service.Version)
	if err := store.DefaultBlobStore.Delete(buildKey, opt); err != nil && err != store.ErrNotFound {
		logger.Warnf("Error deleting build %v: %v", srcKey, err)
	}
}
