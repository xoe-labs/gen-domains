// Code generated by 'ddd-gen app command': DO NOT EDIT.

package command

import (
	"context"
	errwrap "github.com/hashicorp/errwrap"
	app "github.com/xoe-labs/ddd-gen/internal/test-svc/app"
	errors "github.com/xoe-labs/ddd-gen/internal/test-svc/app/errors"
	domain "github.com/xoe-labs/ddd-gen/internal/test-svc/domain"
	"reflect"
)

// Topic: Account

var (
	// ErrNotAuthorizedToArchiveAccount signals that the caller is not authorized to perform ArchiveAccount
	ErrNotAuthorizedToArchiveAccount = errors.NewAuthorizationError("ErrNotAuthorizedToArchiveAccount")
	// ErrArchiveAccountHasNoTarget signals that ArchiveAccount's target was not distinguishable
	ErrArchiveAccountHasNoTarget = errors.NewTargetIdentificationError("ErrArchiveAccountHasNoTarget")
	// ErrArchiveAccountLoadingFailed signals that ArchiveAccount storage failed to load the entity
	ErrArchiveAccountLoadingFailed = errors.NewStorageLoadingError("ErrArchiveAccountLoadingFailed")
	// ErrArchiveAccountSavingFailed signals that ArchiveAccount failed to save the entity
	ErrArchiveAccountSavingFailed = errors.NewStorageSavingError("ErrArchiveAccountSavingFailed")
	// ErrArchiveAccountFailedInDomain signals that ArchiveAccount failed in the domain layer
	ErrArchiveAccountFailedInDomain = errors.NewDomainError("ErrArchiveAccountFailedInDomain")
)

// ArchiveAccountHandlerWrapper knows how to perform ArchiveAccount
type ArchiveAccountHandlerWrapper struct {
	rw app.RequiresStorageWriterReader
	p  app.RequiresPolicer
}

// NewArchiveAccountHandlerWrapper returns ArchiveAccountHandlerWrapper
func NewArchiveAccountHandlerWrapper(rw app.RequiresStorageWriterReader, p app.RequiresPolicer) *ArchiveAccountHandlerWrapper {
	if reflect.ValueOf(rw).IsZero() {
		panic("no 'rw' provided!")
	}
	if reflect.ValueOf(p).IsZero() {
		panic("no 'p' provided!")
	}
	return &ArchiveAccountHandlerWrapper{rw: rw, p: p}
}

// Handle generically performs ArchiveAccount
func (h ArchiveAccountHandlerWrapper) Handle(ctx context.Context, aa domain.ArchiveAccount, actor app.OffersAuthorizable, target app.OffersDistinguishable) error {
	// assert that target is distinguishable
	if !target.IsDistinguishable() {
		return ErrArchiveAccountHasNoTarget
	}
	// load entity from store; handle + wrap error
	a, loadErr := h.rw.Load(ctx, target)
	if loadErr != nil {
		return errwrap.Wrap(ErrArchiveAccountLoadingFailed, loadErr)
	}
	// assert authorization via policy interface
	if ok := h.p.Can(ctx, actor, "ArchiveAccount", a); !ok {
		// return opaque error: handle potentially sensitive policy errors out-of-band!
		return ErrNotAuthorizedToArchiveAccount
	}
	// assert correct command handling by the domain
	if ok := aa.Handle(ctx, a); !ok {
		var domErr error
		for i, e := range aa.Errors() {
			if i == 0 {
				domErr = e
			} else {
				domErr = errwrap.Wrap(domErr, e)
			}
		}
		return ErrArchiveAccountFailedInDomain
	}
	// save domain facts to storage
	saveErr := h.rw.SaveFacts(ctx, target, app.OffersFactKeeper(&aa))
	if saveErr != nil {
		return errwrap.Wrap(ErrArchiveAccountSavingFailed, saveErr)
	}
	return nil
}

// compile time assertions
var (
	_ app.RequiresCommandHandler = (*domain.ArchiveAccount)(nil)
	_ app.RequiresErrorKeeper    = (*domain.ArchiveAccount)(nil)
	_ app.OffersFactKeeper       = (*domain.ArchiveAccount)(nil)
)