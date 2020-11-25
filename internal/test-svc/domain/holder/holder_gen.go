// Code generated by 'ddd-gen domain entity', DO NOT EDIT.
package holder

import (
	"errors"
	"fmt"
	"reflect"
	"time"
)

// Constructors ...

// New returns a guaranteed-to-be-valid Holder or an error
func New(uuid string, name string, bday time.Time, hTyp HolderType) (*Holder, error) {
	if reflect.ValueOf(uuid).IsZero() {
		return nil, errors.New("field uuid is empty")
	}
	if reflect.ValueOf(name).IsZero() {
		return nil, errors.New("field name is empty")
	}
	if reflect.ValueOf(hTyp).IsZero() {
		return nil, errors.New("filed folder type is empty")
	}
	h := &Holder{
		bday: bday,
		hTyp: hTyp,
		name: name,
		uuid: uuid,
	}
	if err := h.validate(); err != nil {
		return nil, err
	}
	return h, nil
}

// MustNew returns a guaranteed-to-be-valid Holder or panics
func MustNew(uuid string, name string, bday time.Time, hTyp HolderType) *Holder {
	h, err := New(uuid, name, bday, hTyp)
	if err != nil {
		panic(err)
	}
	return h
}

// Marshalers ...

// UnmarshalFromStore unmarshals Holder from the repository so that non-constructable
// private fields can still be initialized from (private) repository state
//
// Important: DO NEVER USE THIS METHOD EXCEPT FROM THE REPOSITORY
// Reason: This method initializes private state, so you could corrupt the domain.
func UnmarshalFromStore(uuid string, name string, bday time.Time, hTyp HolderType) *Holder {
	h := MustNew(uuid, name, bday, hTyp)
	return h
}

// Accessors ...

// Utilities ...

// Equal answers whether v is equivalent to h
// Always returns false if v is not a Holder
func (h Holder) Equal(v interface{}) bool {
	other, ok := v.(Holder)
	if !ok {
		return false
	}
	if !reflect.DeepEqual(h.uuid, other.uuid) {
		return false
	}
	return h
}

// String implements the fmt.Stringer interface and returns the native format of Holder
func (h Holder) String() string {
	return fmt.Sprintf("%s ", h.name)
}
