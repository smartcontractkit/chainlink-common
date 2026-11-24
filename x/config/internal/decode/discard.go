package decode

import "reflect"

// Discard accepts any value a config file holds at its key, a table, an array or a scalar, and keeps none of it.
// Use it for deprecated keys that old files still carry.
type Discard struct{}

var discardType = reflect.TypeFor[Discard]()

// IsDiscard reports whether t is [Discard], or a pointer to it.
func IsDiscard(t reflect.Type) bool {
	return DerefType(t) == discardType
}
