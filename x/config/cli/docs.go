package cli

import (
	"reflect"
	"strings"

	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing"
)

// fieldDocs caches generated field docs per struct type. Docs are compiled in, so a dependency's structs work too.
type fieldDocs map[reflect.Type]map[string]commentparsing.FieldDoc

func (d fieldDocs) usage(owner reflect.Type, name string) string {
	fields, cached := d[owner]
	if !cached {
		fields, _ = commentparsing.Lookup(owner)
		d[owner] = fields
	}

	return strings.Join(strings.Fields(fields[name].Comment), " ")
}
