package cli

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/micro/cli/v2"
	"github.com/micro/micro/v3/client/cli/namespace"
	"github.com/micro/micro/v3/client/cli/util"
	pb "github.com/micro/micro/v3/service/auth/proto"
	"github.com/micro/micro/v3/service/errors"
)

func listRules(ctx *cli.Context) error {
	client := pb.NewRulesService("go.micro.auth")

	ns, err := namespace.Get(util.GetEnv(ctx).Name)
	if err != nil {
		return fmt.Errorf("Error getting namespace: %v", err)
	}

	rsp, err := client.List(context.TODO(), &pb.ListRequest{
		Options: &pb.Options{Namespace: ns},
	})
	if err != nil {
		return fmt.Errorf("Error listing rules: %v", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	defer w.Flush()

	formatResource := func(r *pb.Resource) string {
		return strings.Join([]string{r.Type, r.Name, r.Endpoint}, ":")
	}

	// sort rules using resource name and priority to keep the list consistent
	sort.Slice(rsp.Rules, func(i, j int) bool {
		resI := formatResource(rsp.Rules[i].Resource) + string(rsp.Rules[i].Priority)
		resJ := formatResource(rsp.Rules[j].Resource) + string(rsp.Rules[j].Priority)
		return sort.StringsAreSorted([]string{resJ, resI})
	})

	fmt.Fprintln(w, strings.Join([]string{"ID", "Scope", "Access", "Resource", "Priority"}, "\t\t"))
	for _, r := range rsp.Rules {
		res := formatResource(r.Resource)
		if r.Scope == "" {
			r.Scope = "<public>"
		}
		fmt.Fprintln(w, strings.Join([]string{r.Id, r.Scope, r.Access.String(), res, fmt.Sprintf("%d", r.Priority)}, "\t\t"))
	}

	return nil
}

func createRule(ctx *cli.Context) error {
	ns, err := namespace.Get(util.GetEnv(ctx).Name)
	if err != nil {
		return fmt.Errorf("Error getting namespace: %v", err)
	}

	rule, err := constructRule(ctx)
	if err != nil {
		return err
	}

	_, err = pb.NewRulesService("go.micro.auth").Create(context.TODO(), &pb.CreateRequest{
		Rule: rule, Options: &pb.Options{Namespace: ns},
	})
	if verr := errors.Parse(err); verr != nil {
		return fmt.Errorf("Error: %v", verr.Detail)
	} else if err != nil {
		return err
	}

	fmt.Println("Rule created")
	return nil
}

func deleteRule(ctx *cli.Context) error {
	if ctx.Args().Len() != 1 {
		return fmt.Errorf("Expected one argument: ID")
	}

	ns, err := namespace.Get(util.GetEnv(ctx).Name)
	if err != nil {
		return fmt.Errorf("Error getting namespace: %v", err)
	}

	_, err = pb.NewRulesService("go.micro.auth").Delete(context.TODO(), &pb.DeleteRequest{
		Id: ctx.Args().First(), Options: &pb.Options{Namespace: ns},
	})
	if verr := errors.Parse(err); verr != nil {
		return fmt.Errorf("Error: %v", verr.Detail)
	} else if err != nil {
		return err
	}

	fmt.Println("Rule deleted")
	return nil
}

func constructRule(ctx *cli.Context) (*pb.Rule, error) {
	if ctx.Args().Len() != 1 {
		return nil, fmt.Errorf("Too many arguments, expected one argument: ID")
	}

	var access pb.Access
	switch ctx.String("access") {
	case "granted":
		access = pb.Access_GRANTED
	case "denied":
		access = pb.Access_DENIED
	default:
		return nil, fmt.Errorf("Invalid access: %v, must be granted or denied", ctx.String("access"))
	}

	resComps := strings.Split(ctx.String("resource"), ":")
	if len(resComps) != 3 {
		return nil, fmt.Errorf("Invalid resource, must be in the format type:name:endpoint")
	}

	return &pb.Rule{
		Id:       ctx.Args().First(),
		Access:   access,
		Scope:    ctx.String("scope"),
		Priority: int32(ctx.Int("priority")),
		Resource: &pb.Resource{
			Type:     resComps[0],
			Name:     resComps[1],
			Endpoint: resComps[2],
		},
	}, nil
}
