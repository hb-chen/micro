// +build integration

package test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/micro/micro/v3/client/cli/namespace"
	"github.com/micro/micro/v3/client/cli/token"
	"github.com/micro/micro/v3/internal/config"
)

func TestServerAuth(t *testing.T) {
	TrySuite(t, ServerAuth, retryCount)
}

func ServerAuth(t *T) {
	t.Parallel()
	serv := NewServer(t, WithLogin())
	defer serv.Close()
	if err := serv.Run(); err != nil {
		return
	}

	cmd := serv.Command()

	// Execute first command in read to wait for store service
	// to start up
	if err := Try("Calling micro auth list accounts", t, func() ([]byte, error) {
		outp, err := cmd.Exec("auth", "list", "accounts")
		if err != nil {
			return outp, err
		}
		if !strings.Contains(string(outp), "admin") {
			return outp, fmt.Errorf("Output should contain default admin account")
		}
		return outp, nil
	}, 15*time.Second); err != nil {
		return
	}

	if err := Try("Calling micro auth list rules", t, func() ([]byte, error) {
		outp, err := cmd.Exec("auth", "list", "rules")
		if err != nil {
			return outp, err
		}
		if !strings.Contains(string(outp), "default") {
			return outp, fmt.Errorf("Output should contain default rule")
		}
		return outp, nil
	}, 8*time.Second); err != nil {
		return
	}

	if err := Try("Try to get token with default account", t, func() ([]byte, error) {
		outp, err := cmd.Exec("call", "auth", "Auth.Token", `{"id":"admin","secret":"micro"}`)
		if err != nil {
			return outp, err
		}
		rsp := map[string]interface{}{}
		err = json.Unmarshal(outp, &rsp)
		token, ok := rsp["token"].(map[string]interface{})
		if !ok {
			return outp, errors.New("Can't find token")
		}
		if _, ok = token["access_token"].(string); !ok {
			return outp, fmt.Errorf("Can't find access token")
		}
		if _, ok = token["refresh_token"].(string); !ok {
			return outp, fmt.Errorf("Can't find access token")
		}
		if _, ok = token["refresh_token"].(string); !ok {
			return outp, fmt.Errorf("Can't find refresh token")
		}
		if _, ok = token["expiry"].(string); !ok {
			return outp, fmt.Errorf("Can't find access token")
		}
		return outp, nil
	}, 8*time.Second); err != nil {
		return
	}
}

func TestServerLockdown(t *testing.T) {
	TrySuite(t, testServerLockdown, retryCount)
}

func testServerLockdown(t *T) {
	t.Parallel()
	serv := NewServer(t)
	defer serv.Close()
	if err := serv.Run(); err != nil {
		return
	}

	lockdownSuite(serv, t)
}

func lockdownSuite(serv Server, t *T) {
	cmd := serv.Command()

	// Execute first command in read to wait for store service
	// to start up
	ns, err := namespace.Get(serv.Env())
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Namespace is", ns)

	rsp, _ := curl(serv, "store/list")
	if rsp == nil {
		t.Fatal(rsp, errors.New("store list should have response"))
	}
	if val, ok := rsp["Code"].(float64); !ok || val != 401 {
		t.Fatal(rsp, errors.New("store list should be closed"), val)
	}

	Login(serv, t, "admin", "micro")

	email := "me@email.com"
	pass := "mystrongpass"

	outp, err := cmd.Exec("auth", "create", "account", "--secret", pass, email)
	if err != nil {
		t.Fatal(string(outp), err)
		return
	}

	outp, err = cmd.Exec("auth", "create", "rule", "--access=granted", "--scope=*", "--resource=*:*:*", "onlyloggedin")
	if err != nil {
		t.Fatal(string(outp), err)
		return
	}

	outp, err = cmd.Exec("auth", "create", "rule", "--access=granted", "--resource=service:auth:Auth.Token", "authpublic")
	if err != nil {
		t.Fatal(string(outp), err)
		return
	}

	outp, err = cmd.Exec("auth", "delete", "rule", "admin")
	if err != nil {
		t.Fatal(string(outp), err)
		return
	}

	outp, err = cmd.Exec("auth", "delete", "account", "admin")
	if err != nil {
		t.Fatal(string(outp), err)
		return
	}

	// set the local config file to be the same as the one micro will be configured to use.
	// todo: consider adding a micro logout command.
	config.SetConfig(cmd.Config)
	err = token.Remove(serv.Env())
	if err != nil {
		t.Fatal(err)
		return
	}

	if err := Try("Listing rules should fail before login", t, func() ([]byte, error) {
		outp, err := cmd.Exec("auth", "list", "rules")
		if err == nil {
			return outp, errors.New("List rules should fail")
		}
		return outp, nil
	}, 40*time.Second); err != nil {
		return
	}

	// auth rules are cached so this could take a few seconds (until the authpublic rule takes
	// effect in both the proxy and the auth service)
	if err := Try("Logging in with "+email, t, func() ([]byte, error) {
		out, err := cmd.Exec("login", "--email", email, "--password", pass)
		if err != nil {
			return out, err
		}
		if !strings.Contains(string(out), "Success") {
			return out, errors.New("Login output does not contain 'Success'")
		}
		return out, err
	}, 40*time.Second); err != nil {
		return
	}

	if err := Try("Listing rules should pass after login", t, func() ([]byte, error) {
		outp, err := cmd.Exec("auth", "list", "rules")
		if err != nil {
			return outp, err
		}
		if !strings.Contains(string(outp), "onlyloggedin") || !strings.Contains(string(outp), "authpublic") {
			return outp, errors.New("Can't find rules")
		}
		return outp, err
	}, 40*time.Second); err != nil {
		return
	}
}

func TestPasswordChange(t *testing.T) {
	TrySuite(t, changePassword, retryCount)
}

func changePassword(t *T) {
	t.Parallel()
	serv := NewServer(t, WithLogin())
	defer serv.Close()
	if err := serv.Run(); err != nil {
		return
	}

	cmd := serv.Command()
	newPass := "shinyNewPass"

	// Bad password should not succeed
	outp, err := cmd.Exec("user", "set", "password", "--old-password", "micro121212", "--new-password", newPass)
	if err == nil {
		t.Fatal("Incorrect existing password should make password change fail")
		return
	}

	outp, err = cmd.Exec("user", "set", "password", "--old-password", "micro", "--new-password", newPass)
	if err != nil {
		t.Fatal(string(outp))
		return
	}

	time.Sleep(3 * time.Second)
	outp, err = cmd.Exec("login", "--email", "admin", "--password", "micro")
	if err == nil {
		t.Fatal("Old password should not be usable anymore")
		return
	}
	outp, err = cmd.Exec("login", "--email", "admin", "--password", newPass)
	if err != nil {
		t.Fatal(string(outp))
		return
	}
}

func curl(serv Server, path string) (map[string]interface{}, error) {
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%v/%v", serv.APIPort(), path))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	m := map[string]interface{}{}
	return m, json.Unmarshal(body, &m)
}
