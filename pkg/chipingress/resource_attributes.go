package chipingress

import "sort"

// resourceAttrKey pairs a sanitized extension/metadata key name with the original
// resource-attribute key it was derived from.
type resourceAttrKey struct {
	name string
	key  string
}

// sanitizeResourceAttributeKeys returns the deduplicated, sorted list of resource-attribute
// keys that survive sanitization and reservation checks. The returned pairs contain the
// sanitized name and the original map key, so callers can apply their own value handling.
//
// Ordering is deterministic: original keys are sorted lexicographically, and if two keys
// sanitize to the same name the first one in sorted order wins. extraReserved, if non-nil,
// is consulted in addition to reservedExtensionNames.
func sanitizeResourceAttributeKeys(attrs map[string]string, extraReserved map[string]struct{}) []resourceAttrKey {
	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	seen := make(map[string]struct{}, len(attrs))
	result := make([]resourceAttrKey, 0, len(attrs))
	for _, k := range keys {
		name := sanitizeExtensionName(k)
		if name == "" {
			continue
		}
		if _, reserved := reservedExtensionNames[name]; reserved {
			continue
		}
		if _, reserved := extraReserved[name]; reserved {
			continue
		}
		if _, already := seen[name]; already {
			continue
		}
		seen[name] = struct{}{}
		result = append(result, resourceAttrKey{name: name, key: k})
	}
	return result
}
