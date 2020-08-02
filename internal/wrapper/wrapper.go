package wrapper

import (
	"context"
	"reflect"
	"strings"

	goauth "github.com/micro/go-micro/v3/auth"
	"github.com/micro/go-micro/v3/client"
	"github.com/micro/go-micro/v3/debug/trace"
	"github.com/micro/go-micro/v3/metadata"
	"github.com/micro/go-micro/v3/server"
	"github.com/micro/micro/v3/internal/namespace"
	"github.com/micro/micro/v3/service/auth"
	muclient "github.com/micro/micro/v3/service/client"
	"github.com/micro/micro/v3/service/debug"
	"github.com/micro/micro/v3/service/errors"
	muserver "github.com/micro/micro/v3/service/server"
)

type authWrapper struct {
	client.Client
}

func (a *authWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	ctx = a.wrapContext(ctx, opts...)
	return a.Client.Call(ctx, req, rsp, opts...)
}

func (a *authWrapper) Stream(ctx context.Context, req client.Request, opts ...client.CallOption) (client.Stream, error) {
	ctx = a.wrapContext(ctx, opts...)
	return a.Client.Stream(ctx, req, opts...)
}

func (a *authWrapper) wrapContext(ctx context.Context, opts ...client.CallOption) context.Context {
	// parse the options
	var options client.CallOptions
	for _, o := range opts {
		o(&options)
	}

	// check to see if the authorization header has already been set.
	// We dont't override the header unless the AuthToken option has
	// been specified or the header wasn't provided
	if _, ok := metadata.Get(ctx, "Authorization"); ok && !options.AuthToken {
		return ctx
	}

	// set the namespace header if it has not been set (e.g. on a service to service request)
	authOpts := auth.DefaultAuth.Options()
	if _, ok := metadata.Get(ctx, "Micro-Namespace"); !ok {
		ctx = metadata.Set(ctx, "Micro-Namespace", authOpts.Issuer)
	}

	// check to see if we have a valid access token
	if authOpts.Token != nil && !authOpts.Token.Expired() {
		ctx = metadata.Set(ctx, "Authorization", goauth.BearerScheme+authOpts.Token.AccessToken)
		return ctx
	}

	// call without an auth token
	return ctx
}

// AuthClient wraps requests with the auth header
func AuthClient(c client.Client) client.Client {
	return &authWrapper{c}
}

// AuthHandler wraps a server handler to perform auth
func AuthHandler() server.HandlerWrapper {
	return func(h server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			// Extract the token if the header is present. We will inspect the token regardless of if it's
			// present or not since noop auth will return a blank account upon Inspecting a blank token.
			var token string
			if header, ok := metadata.Get(ctx, "Authorization"); ok {
				// Ensure the correct scheme is being used
				if !strings.HasPrefix(header, goauth.BearerScheme) {
					return errors.Unauthorized(req.Service(), "invalid authorization header. expected Bearer schema")
				}

				// Strip the bearer scheme prefix
				token = strings.TrimPrefix(header, goauth.BearerScheme)
			}

			// Inspect the token and decode an account
			account, _ := auth.Inspect(token)

			// ensure only accounts with the correct namespace can access this namespace,
			// since the auth package will verify access below, and some endpoints could
			// be public, we allow nil accounts access using the namespace.Public option.
			ns := auth.DefaultAuth.Options().Issuer
			err := namespace.Authorize(ctx, ns, namespace.Public(ns))
			if err == namespace.ErrForbidden {
				return errors.Forbidden(req.Service(), err.Error())
			} else if err != nil {
				return errors.InternalServerError(req.Service(), err.Error())
			}

			// construct the resource
			res := &goauth.Resource{
				Type:     "service",
				Name:     req.Service(),
				Endpoint: req.Endpoint(),
			}

			// Verify the caller has access to the resource.
			err = auth.Verify(account, res, goauth.VerifyNamespace(ns))
			if err == goauth.ErrForbidden && account != nil {
				return errors.Forbidden(req.Service(), "Forbidden call made to %v:%v by %v", req.Service(), req.Endpoint(), account.ID)
			} else if err == goauth.ErrForbidden {
				return errors.Unauthorized(req.Service(), "Unauthorized call made to %v:%v", req.Service(), req.Endpoint())
			} else if err != nil {
				return errors.InternalServerError(req.Service(), "Error authorizing request: %v", err)
			}

			// There is an account, set it in the context
			if account != nil {
				ctx = goauth.ContextWithAccount(ctx, account)
			}

			// The user is authorised, allow the call
			return h(ctx, req, rsp)
		}
	}
}

type fromServiceWrapper struct {
	client.Client
}

var (
	HeaderPrefix = "Micro-"
)

func (f *fromServiceWrapper) setHeaders(ctx context.Context) context.Context {
	return metadata.MergeContext(ctx, metadata.Metadata{
		HeaderPrefix + "From-Service": muserver.DefaultServer.Options().Name,
	}, false)
}

func (f *fromServiceWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	ctx = f.setHeaders(ctx)
	return f.Client.Call(ctx, req, rsp, opts...)
}

func (f *fromServiceWrapper) Stream(ctx context.Context, req client.Request, opts ...client.CallOption) (client.Stream, error) {
	ctx = f.setHeaders(ctx)
	return f.Client.Stream(ctx, req, opts...)
}

func (f *fromServiceWrapper) Publish(ctx context.Context, p client.Message, opts ...client.PublishOption) error {
	ctx = f.setHeaders(ctx)
	return f.Client.Publish(ctx, p, opts...)
}

// FromService wraps a client to inject service and auth metadata
func FromService(c client.Client) client.Client {
	return &fromServiceWrapper{c}
}

// HandlerStats wraps a server handler to generate request/error stats
func HandlerStats() server.HandlerWrapper {
	// return a handler wrapper
	return func(h server.HandlerFunc) server.HandlerFunc {
		// return a function that returns a function
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			// execute the handler
			err := h(ctx, req, rsp)
			// record the stats
			debug.DefaultStats.Record(err)
			// return the error
			return err
		}
	}
}

type traceWrapper struct {
	client.Client
}

func (c *traceWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	newCtx, s := debug.DefaultTracer.Start(ctx, req.Service()+"."+req.Endpoint())

	s.Type = trace.SpanTypeRequestOutbound
	err := c.Client.Call(newCtx, req, rsp, opts...)
	if err != nil {
		s.Metadata["error"] = err.Error()
	}

	// finish the trace
	debug.DefaultTracer.Finish(s)

	return err
}

// TraceCall is a call tracing wrapper
func TraceCall(c client.Client) client.Client {
	return &traceWrapper{
		Client: c,
	}
}

// TraceHandler wraps a server handler to perform tracing
func TraceHandler() server.HandlerWrapper {
	// return a handler wrapper
	return func(h server.HandlerFunc) server.HandlerFunc {
		// return a function that returns a function
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			// don't store traces for debug
			if strings.HasPrefix(req.Endpoint(), "Debug.") {
				return h(ctx, req, rsp)
			}

			// get the span
			newCtx, s := debug.DefaultTracer.Start(ctx, req.Service()+"."+req.Endpoint())
			s.Type = trace.SpanTypeRequestInbound

			err := h(newCtx, req, rsp)
			if err != nil {
				s.Metadata["error"] = err.Error()
			}

			// finish
			debug.DefaultTracer.Finish(s)

			return err
		}
	}
}

type cacheWrapper struct {
	client.Client
}

// Call executes the request. If the CacheExpiry option was set, the response will be cached using
// a hash of the metadata and request as the key.
func (c *cacheWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	// parse the options
	var options client.CallOptions
	for _, o := range opts {
		o(&options)
	}

	// if the client doesn't have a cacbe setup don't continue
	cache := muclient.DefaultClient.Options().Cache
	if cache == nil {
		return c.Client.Call(ctx, req, rsp, opts...)
	}

	// if the cache expiry is not set, execute the call without the cache
	if options.CacheExpiry == 0 {
		return c.Client.Call(ctx, req, rsp, opts...)
	}

	// if the response is nil don't call the cache since we can't assign the response
	if rsp == nil {
		return c.Client.Call(ctx, req, rsp, opts...)
	}

	// check to see if there is a response cached, if there is assign it
	if r, ok := cache.Get(ctx, req); ok {
		val := reflect.ValueOf(rsp).Elem()
		val.Set(reflect.ValueOf(r).Elem())
		return nil
	}

	// don't cache the result if there was an error
	if err := c.Client.Call(ctx, req, rsp, opts...); err != nil {
		return err
	}

	// set the result in the cache
	cache.Set(ctx, req, rsp, options.CacheExpiry)
	return nil
}

// CacheClient wraps requests with the cache wrapper
func CacheClient(c client.Client) client.Client {
	return &cacheWrapper{c}
}
