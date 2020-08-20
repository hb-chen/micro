package auth

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/micro/go-micro/v3/auth"
	"github.com/micro/go-micro/v3/store"
	gostore "github.com/micro/go-micro/v3/store"
	"github.com/micro/micro/v3/internal/namespace"
	pb "github.com/micro/micro/v3/service/auth/proto"
	"github.com/micro/micro/v3/service/errors"
)

// List returns all auth accounts
func (a *Auth) List(ctx context.Context, req *pb.ListAccountsRequest, rsp *pb.ListAccountsResponse) error {
	// set defaults
	if req.Options == nil {
		req.Options = &pb.Options{}
	}
	if len(req.Options.Namespace) == 0 {
		req.Options.Namespace = namespace.DefaultNamespace
	}

	// authorize the request
	if err := namespace.Authorize(ctx, req.Options.Namespace); err == namespace.ErrForbidden {
		return errors.Forbidden("auth.Accounts.List", err.Error())
	} else if err == namespace.ErrUnauthorized {
		return errors.Unauthorized("auth.Accounts.List", err.Error())
	} else if err != nil {
		return errors.InternalServerError("auth.Accounts.List", err.Error())
	}

	// setup the defaults incase none exist
	a.setupDefaultAccount(req.Options.Namespace)

	// get the records from the store
	key := strings.Join([]string{storePrefixAccounts, req.Options.Namespace, ""}, joinKey)
	recs, err := a.Options.Store.Read(key, store.ReadPrefix())
	if err != nil {
		return errors.InternalServerError("auth.Accounts.List", "Unable to read from store: %v", err)
	}

	// unmarshal the records
	var accounts = make([]*auth.Account, 0, len(recs))
	for _, rec := range recs {
		var r *auth.Account
		if err := json.Unmarshal(rec.Value, &r); err != nil {
			return errors.InternalServerError("auth.Accounts.List", "Error to unmarshaling json: %v. Value: %v", err, string(rec.Value))
		}
		accounts = append(accounts, r)
	}

	// serialize the accounts
	rsp.Accounts = make([]*pb.Account, 0, len(recs))
	for _, a := range accounts {
		rsp.Accounts = append(rsp.Accounts, serializeAccount(a))
	}

	return nil
}

// Delete an auth account
func (a *Auth) Delete(ctx context.Context, req *pb.DeleteAccountRequest, rsp *pb.DeleteAccountResponse) error {
	// validate the request
	if len(req.Id) == 0 {
		return errors.BadRequest("auth.Accounts.Delete", "Missing ID")
	}

	// set defaults
	if req.Options == nil {
		req.Options = &pb.Options{}
	}
	if len(req.Options.Namespace) == 0 {
		req.Options.Namespace = namespace.DefaultNamespace
	}

	// authorize the request
	if err := namespace.Authorize(ctx, req.Options.Namespace); err == namespace.ErrForbidden {
		return errors.Forbidden("auth.Accounts.Delete", err.Error())
	} else if err == namespace.ErrUnauthorized {
		return errors.Unauthorized("auth.Accounts.Delete", err.Error())
	} else if err != nil {
		return errors.InternalServerError("auth.Accounts.Delete", err.Error())
	}

	// check the account exists
	key := strings.Join([]string{storePrefixAccounts, req.Options.Namespace, req.Id}, joinKey)
	if _, err := a.Options.Store.Read(key); err == store.ErrNotFound {
		return errors.BadRequest("auth.Accounts.Delete", "Account not found with this ID")
	} else if err != nil {
		return errors.BadRequest("auth.Accounts.Delete", "Error querying accounts: %v", err)
	}

	// delete the refresh token linked to the account
	tok, err := a.refreshTokenForAccount(req.Options.Namespace, req.Id)
	if err != nil {
		return errors.InternalServerError("auth.Accounts.Delete", "Error finding refresh token")
	}
	refreshKey := strings.Join([]string{storePrefixRefreshTokens, req.Options.Namespace, req.Id, tok}, joinKey)
	if err := a.Options.Store.Delete(refreshKey); err != nil {
		return errors.InternalServerError("auth.Accounts.Delete", "Error deleting refresh token: %v", err)
	}

	// delete the account
	if err := a.Options.Store.Delete(key); err != nil {
		return errors.BadRequest("auth.Accounts.Delete", "Error deleting account: %v", err)
	}

	// Clear the namespace cache, since the accounts for this namespace could now be empty
	a.Lock()
	delete(a.namespaces, req.Options.Namespace)
	a.Unlock()

	return nil
}

// ChangeSecret by providing a refresh token and a new secret
func (a *Auth) ChangeSecret(ctx context.Context, req *pb.ChangeSecretRequest, rsp *pb.ChangeSecretResponse) error {
	if len(req.NewSecret) == 0 {
		return errors.BadRequest("auth.Auth.ChangeSecret", "New secret should not be blank")
	}

	// set defaults
	if req.Options == nil {
		req.Options = &pb.Options{}
	}
	if len(req.Options.Namespace) == 0 {
		req.Options.Namespace = namespace.DefaultNamespace
	}

	// authorize the request
	if err := namespace.Authorize(ctx, req.Options.Namespace); err == namespace.ErrForbidden {
		return errors.Forbidden("auth.Accounts.ChangeSecret", err.Error())
	} else if err == namespace.ErrUnauthorized {
		return errors.Unauthorized("auth.Accounts.ChangeSecret", err.Error())
	} else if err != nil {
		return errors.InternalServerError("auth.Accounts.ChangeSecret", err.Error())
	}

	// Lookup the account in the store
	key := strings.Join([]string{storePrefixAccounts, req.Options.Namespace, req.Id}, joinKey)
	recs, err := a.Options.Store.Read(key)
	if err == gostore.ErrNotFound {
		return errors.BadRequest("auth.Auth.ChangeSecret", "Account not found with this ID")
	} else if err != nil {
		return errors.InternalServerError("auth.Accounts.ChangeSecret", "Unable to read from store: %v", err)
	}

	// Unmarshal the record
	var acc *auth.Account
	if err := json.Unmarshal(recs[0].Value, &acc); err != nil {
		return errors.InternalServerError("auth.Accounts.ChangeSecret", "Unable to unmarshal account: %v", err)
	}

	if !secretsMatch(acc.Secret, req.OldSecret) {
		return errors.BadRequest("auth.Accounts.ChangeSecret", "Secret not correct")
	}

	// hash the secret
	secret, err := hashSecret(req.NewSecret)
	if err != nil {
		return errors.InternalServerError("auth.Accounts.ChangeSecret", "Unable to hash password: %v", err)
	}
	acc.Secret = secret

	// marshal to json
	bytes, err := json.Marshal(acc)
	if err != nil {
		return errors.InternalServerError("auth.Accounts.ChangeSecret", "Unable to marshal json: %v", err)
	}

	// write to the store
	if err := a.Options.Store.Write(&gostore.Record{Key: key, Value: bytes}); err != nil {
		return errors.InternalServerError("auth.Accounts.ChangeSecret", "Unable to write account to store: %v", err)
	}
	return nil
}

func serializeAccount(a *auth.Account) *pb.Account {
	return &pb.Account{
		Id:       a.ID,
		Type:     a.Type,
		Scopes:   a.Scopes,
		Issuer:   a.Issuer,
		Metadata: a.Metadata,
	}
}
