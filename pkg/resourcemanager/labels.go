package resourcemanager

import "strings"

// OwnerLabel returns the canonical form of the "owner" label, used by all
// producers: a leading "0x"/"0X" prefix is stripped and the remainder is
// lowercased. Downstream joins on the owner label depend on every producer
// emitting exactly this form, so producers must not write the owner label
// any other way.
func OwnerLabel(owner string) string {
	if strings.HasPrefix(owner, "0x") || strings.HasPrefix(owner, "0X") {
		owner = owner[2:]
	}
	return strings.ToLower(owner)
}
