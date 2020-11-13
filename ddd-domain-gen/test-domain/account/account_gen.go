// Code generated by ddd-domain-gen, DO NOT EDIT.
package account

import (
	"errors"
	holder "github.com/xoe-labs/go-generators/ddd-domain-gen/test-domain/holder"
)

// Generators ...

// New returns a guaranteed-to-be-valid Account or an error
func New(uuid *string, holder *holder.Holder, address *string) (*Account, error) {
	if uuid == nil {
		return nil, errors.New("field uuid is missing")
	}
	if holder == nil {
		return nil, errors.New("field holder is missing")
	}
	return &Account{
		address: address,
		holder:  holder,
		uuid:    uuid,
	}, nil
}

// MustNew returns a guaranteed-to-be-valid Account or panics
func MustNew(uuid *string, holder *holder.Holder, address *string) *Account {
	a, err := New(uuid, holder, address)
	if err != nil {
		panic(err)
	}
	return a
}

// Marshalers ...

// UnmarshalFromRepository unmarshals Account from the repository so that non-constructable
// private fields can still be initialized from (private) repository state
//
// Important: DO NEVER USE THIS METHOD EXCEPT FROM THE REPOSITORY
// Reason: This method initializes private state, so you could corrupt the domain.
func UnmarshalFromRepository(uuid *string, holder *holder.Holder, address *string, balance *int64, values *[]int64) *Account {
	a := MustNew(uuid, holder, address)
	a.balance = balance
	a.values = values
	return a
}
