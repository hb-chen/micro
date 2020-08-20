package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/micro/cli/v2"
	goclient "github.com/micro/go-micro/v3/client"
	goregistry "github.com/micro/go-micro/v3/registry"
	"github.com/micro/micro/v3/client/cli/namespace"
	"github.com/micro/micro/v3/client/cli/util"
	"github.com/micro/micro/v3/service/client"
	"github.com/micro/micro/v3/service/registry"
)

// lookupService queries the service for a service with the given alias. If
// no services are found for a given alias, the registry will return nil and
// the error will also be nil. An error is only returned if there was an issue
// listing from the registry.
func lookupService(ctx *cli.Context) (*goregistry.Service, error) {
	// use the first arg as the name, e.g. "micro helloworld foo"
	// would try to call the helloworld service
	name := ctx.Args().First()

	// get the namespace to query the services from
	dom, err := namespace.Get(util.GetEnv(ctx).Name)
	if err != nil {
		return nil, err
	}

	// lookup from the registry in the current namespace
	if srv, err := serviceWithName(name, dom); err != nil {
		return nil, err
	} else if srv != nil {
		return srv, nil
	}

	// check for the service in the default namespace also
	if dom == goregistry.DefaultDomain {
		return nil, nil
	}
	return serviceWithName(name, goregistry.DefaultDomain)
}

// formatServiceUsage returns a string containing the service usage.
func formatServiceUsage(srv *goregistry.Service, alias string) string {
	commands := make([]string, len(srv.Endpoints))
	for i, e := range srv.Endpoints {
		// map "Helloworld.Call" to "helloworld.call"
		name := strings.ToLower(e.Name)

		// remove the prefix if it is the service name, e.g. rather than
		// "micro run helloworld helloworld call", it would be
		// "micro run helloworld call".
		name = strings.TrimPrefix(name, alias+".")

		// instead of "micro run helloworld foo.bar", the command should
		// be "micro run helloworld foo bar".
		commands[i] = strings.Replace(name, ".", " ", 1)
	}

	// sort the command names alphabetically
	sort.Strings(commands)

	result := fmt.Sprintf("NAME:\n\t%v\n\n", srv.Name)
	result += fmt.Sprintf("VERSION:\n\t%v\n\n", srv.Version)
	result += fmt.Sprintf("USAGE:\n\tmicro %v [flags] [command]\n\n", alias)
	result += fmt.Sprintf("COMMANDS:\n\t%v\n\n", strings.Join(commands, "\n\t"))
	return result
}

// callService will call a service using the arguments and flags provided
// in the context. It will print the result or error to stdout. If there
// was an error performing the call, it will be returned.
func callService(srv *goregistry.Service, ctx *cli.Context) error {
	// parse the flags and args
	args, flags, err := splitCmdArgs(ctx)
	if err != nil {
		return err
	}

	// construct the endpoint
	endpoint, err := constructEndpoint(args)
	if err != nil {
		return err
	}

	// ensure the endpoint exists on the service
	var ep *goregistry.Endpoint
	for _, e := range srv.Endpoints {
		if e.Name == endpoint {
			ep = e
			break
		}
	}
	if ep == nil {
		return fmt.Errorf("Endpoint %v not found for service %v", endpoint, srv.Name)
	}

	// parse the flags
	body, err := flagsToRequest(flags, ep.Request)
	if err != nil {
		return err
	}

	// construct and execute the request using the json content type
	req := client.NewRequest(srv.Name, endpoint, body, goclient.WithContentType("application/json"))
	var rsp json.RawMessage
	if err := client.Call(ctx.Context, req, &rsp, goclient.WithAuthToken()); err != nil {
		return err
	}

	// format the response
	var out bytes.Buffer
	defer out.Reset()
	if err := json.Indent(&out, rsp, "", "\t"); err != nil {
		return err
	}
	out.Write([]byte("\n"))
	out.WriteTo(os.Stdout)

	return nil
}

// splitCmdArgs takes a cli context and parses out the args and flags, for
// example "micro helloworld --name=foo call apple" would result in "call",
// "apple" as args and {"name":"foo"} as the flags.
func splitCmdArgs(ctx *cli.Context) ([]string, map[string]string, error) {
	args := []string{}
	flags := map[string]string{}

	for _, a := range ctx.Args().Slice() {
		if !strings.HasPrefix(a, "--") {
			args = append(args, a)
			continue
		}

		// comps would be "foo", "bar" for "--foo=bar"
		comps := strings.Split(strings.TrimPrefix(a, "--"), "=")
		switch len(comps) {
		case 1:
			flags[comps[0]] = ""
		case 2:
			flags[comps[0]] = comps[1]
		default:
			return nil, nil, fmt.Errorf("Invalid flag: %v. Expected format: --foo=bar", a)
		}
	}

	return args, flags, nil
}

// constructEndpoint takes a slice of args and converts it into a valid endpoint
// such as Helloworld.Call or Foo.Bar, it will return an error if an invalid number
// of arguments were provided
func constructEndpoint(args []string) (string, error) {
	var epComps []string
	switch len(args) {
	case 1:
		epComps = append(args, "call")
	case 2:
		epComps = args
	case 3:
		epComps = args[1:3]
	default:
		return "", fmt.Errorf("Incorrect number of arguments")
	}

	// transform the endpoint components, e.g ["helloworld", "call"] to the
	// endpoint name: "Helloworld.Call".
	return fmt.Sprintf("%v.%v", strings.Title(epComps[0]), strings.Title(epComps[1])), nil
}

// shouldRenderHelp returns true if the help flag was passed
func shouldRenderHelp(ctx *cli.Context) bool {
	_, flags, _ := splitCmdArgs(ctx)
	for key := range flags {
		if key == "help" {
			return true
		}
	}
	return false
}

// flagsToRequeest parses a set of flags, e.g {name:"Foo", "options_surname","Bar"} and
// converts it into a request body. If the key is not a valid object in the request, an
// error will be returned.
func flagsToRequest(flags map[string]string, req *goregistry.Value) (map[string]interface{}, error) {
	result := map[string]interface{}{}

loop:
	for key, value := range flags {
		for _, attr := range req.Values {
			// matches at a top level
			if attr.Name == key {
				result[key] = value
				continue loop
			}

			// check for matches at the second level
			if !strings.HasPrefix(key, attr.Name+"_") {
				continue
			}
			for _, attr2 := range attr.Values {
				if attr.Name+"_"+attr2.Name != key {
					continue
				}

				if _, ok := result[attr.Name]; !ok {
					result[attr.Name] = map[string]string{}
				} else if _, ok := result[attr.Name].(map[string]string); !ok {
					return nil, fmt.Errorf("Error parsing request, duplicate key: %v", key)
				}
				result[attr.Name].(map[string]string)[attr2.Name] = value
				continue loop
			}
		}

		return nil, fmt.Errorf("Unknown flag: %v", key)
	}

	return result, nil
}

// find a service in a domain matching the name
func serviceWithName(name, domain string) (*goregistry.Service, error) {
	srvs, err := registry.GetService(name, goregistry.GetDomain(domain))
	if err == goregistry.ErrNotFound {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	if len(srvs) == 0 {
		return nil, nil
	}
	return srvs[0], nil
}
