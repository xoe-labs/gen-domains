// Code generated by 'ddd-gen app command': DO NOT EDIT.

package command

import (
	"context"
	"encoding/json"
	errwrap "github.com/hashicorp/errwrap"
	error1 "github.com/xoe-labs/ddd-gen/internal/test-svc/app/error"
	policy "github.com/xoe-labs/ddd-gen/internal/test-svc/app/policy"
	repository "github.com/xoe-labs/ddd-gen/internal/test-svc/app/repository"
	account "github.com/xoe-labs/ddd-gen/internal/test-svc/domain/account"
	"reflect"
)

// Topic: Balance

var (
	// ErrNotAuthorizedToIncreaseBalance signals that the caller is not authorized to perform IncreaseBalance
	ErrNotAuthorizedToIncreaseBalance = error1.NewAuthorizationError("ErrNotAuthorizedToIncreaseBalance")
	// ErrIncreaseBalanceNotIdentifiable signals that IncreaseBalance's command object was not identifiable
	ErrIncreaseBalanceNotIdentifiable = error1.NewIdentificationError("ErrIncreaseBalanceNotIdentifiable")
	// ErrIncreaseBalanceFailedInRepository signals that IncreaseBalance failed in the repository layer
	ErrIncreaseBalanceFailedInRepository = error1.NewRepositoryError("ErrIncreaseBalanceFailedInRepository")
	// ErrIncreaseBalanceFailedInDomain signals that IncreaseBalance failed in the domain layer
	ErrIncreaseBalanceFailedInDomain = error1.NewDomainError("ErrIncreaseBalanceFailedInDomain")
)

// IncreaseBalanceHandler knows how to perform IncreaseBalance
type IncreaseBalanceHandler struct {
	pol policy.Policer
	agg repository.Repository
}

// NewIncreaseBalanceHandler returns IncreaseBalanceHandler
func NewIncreaseBalanceHandler(pol policy.Policer, agg repository.Repository) *IncreaseBalanceHandler {
	if reflect.ValueOf(pol).IsZero() {
		panic("no 'pol' provided!")
	}
	if reflect.ValueOf(agg).IsZero() {
		panic("no 'agg' provided!")
	}
	return &IncreaseBalanceHandler{pol: pol, agg: agg}
}

// Handle generically performs IncreaseBalance
func (h IncreaseBalanceHandler) Handle(ctx context.Context, ib IncreaseBalance) error {
	if reflect.ValueOf(ib.Identifier()).IsZero() {
		return ErrIncreaseBalanceNotIdentifiable
	}
	var innerErr error
	var repoErr error
	repoErr = h.agg.Update(ctx, ib, func(a *account.Account) bool {
		data, err := json.Marshal(a)
		if err != nil {
			panic(err) // invariant violation: the domain shall always be consistent!
		}
		if ok := h.pol.Can(ctx, ib, "IncreaseBalance", data); !ok {
			innerErr = ErrNotAuthorizedToIncreaseBalance
			return false
		}
		if err := ib.handle(ctx, a); err != nil {
			innerErr = errwrap.Wrap(ErrIncreaseBalanceFailedInDomain, err)
			return false
		}
		return true
	})
	if innerErr != nil {
		return innerErr
	}
	if repoErr != nil {
		return errwrap.Wrap(ErrIncreaseBalanceFailedInRepository, repoErr)
	}
	return nil
}