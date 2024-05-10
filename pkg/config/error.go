package config

import (
	"fmt"
	"reflect"
	"strings"

	"go.uber.org/multierr"
)

// lightweight error types copied from core

type ErrInvalid struct {
	Name  string
	Value any
	Msg   string
}

func (e ErrInvalid) Error() string {
	return fmt.Sprintf("%s: invalid value (%v): %s", e.Name, e.Value, e.Msg)
}

// NewErrDuplicate returns an ErrInvalid with a standard duplicate message.
func NewErrDuplicate(name string, value any) ErrInvalid {
	return ErrInvalid{Name: name, Value: value, Msg: "duplicate - must be unique"}
}

type ErrMissing struct {
	Name string
	Msg  string
}

func (e ErrMissing) Error() string {
	return fmt.Sprintf("%s: missing: %s", e.Name, e.Msg)
}

type ErrEmpty struct {
	Name string
	Msg  string
}

func (e ErrEmpty) Error() string {
	return fmt.Sprintf("%s: empty: %s", e.Name, e.Msg)
}

type KeyNotFoundError struct {
	ID      string
	KeyType string
}

func (e KeyNotFoundError) Error() string {
	return fmt.Sprintf("unable to find %s key with id %s", e.KeyType, e.ID)
}

// UniqueStrings is a helper for tracking unique values in string form.
type UniqueStrings map[string]struct{}

// IsDupeFmt is like IsDupe, but calls String().
func (u UniqueStrings) IsDupeFmt(t fmt.Stringer) bool {
	if t == nil {
		return false
	}
	if reflect.ValueOf(t).IsNil() {
		// interface holds a typed-nil value
		return false
	}
	return u.isDupe(t.String())
}

// IsDupe returns true if the set already contains the string, otherwise false.
// Non-nil/empty strings are added to the set.
func (u UniqueStrings) IsDupe(s *string) bool {
	if s == nil {
		return false
	}
	return u.isDupe(*s)
}

func (u UniqueStrings) isDupe(s string) bool {
	if s == "" {
		return false
	}
	_, ok := u[s]
	if !ok {
		u[s] = struct{}{}
	}
	return ok
}

type multiErrorList []error

// MultiErrorList returns an error which formats underlying errors as a list, or nil if err is nil.
func MultiErrorList(err error) (int, error) {
	if err == nil {
		return 0, nil
	}
	errs := multierr.Errors(err)
	return len(errs), multiErrorList(errs)
}

func (m multiErrorList) Error() string {
	l := len(m)
	if l == 1 {
		return m[0].Error()
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "%d errors:", l)
	for _, e := range m {
		fmt.Fprintf(&sb, "\n\t- %v", e)
	}
	return sb.String()
}

func (m multiErrorList) Unwrap() []error {
	return m
}
