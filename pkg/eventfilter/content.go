package eventfilter

import (
	"strconv"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// resolveFieldValue walks a dot-separated field path over msg and returns the
// stringified scalar leaf value. It returns ok=false when any segment is
// missing, an intermediate segment is not a singular message, or the leaf is
// repeated/map (unsupported in v1). No value is ever logged from here.
func resolveFieldValue(msg protoreflect.Message, path string) (string, bool) {
	segs := strings.Split(path, ".")
	cur := msg
	for i, seg := range segs {
		fd := cur.Descriptor().Fields().ByName(protoreflect.Name(seg))
		if fd == nil {
			return "", false
		}
		if i == len(segs)-1 {
			if fd.IsList() || fd.IsMap() {
				return "", false
			}
			return scalarString(fd, cur.Get(fd)), true
		}
		// Intermediate segment must be a singular message to keep traversing.
		if fd.Kind() != protoreflect.MessageKind || fd.IsList() || fd.IsMap() {
			return "", false
		}
		cur = cur.Get(fd).Message()
	}
	return "", false
}

// scalarString renders a scalar field value to a string for regex matching.
func scalarString(fd protoreflect.FieldDescriptor, v protoreflect.Value) string {
	switch fd.Kind() {
	case protoreflect.StringKind:
		return v.String()
	case protoreflect.BytesKind:
		return string(v.Bytes())
	case protoreflect.EnumKind:
		if ev := fd.Enum().Values().ByNumber(v.Enum()); ev != nil {
			return string(ev.Name())
		}
		return strconv.FormatInt(int64(v.Enum()), 10)
	default:
		return v.String()
	}
}

// allTextValues recursively collects every string and bytes leaf value in msg
// (including those nested in messages, repeated fields, and map values), joined
// by newlines, for whole-message ("top-level") content matching.
func allTextValues(msg protoreflect.Message) string {
	var sb strings.Builder
	collectText(msg, &sb)
	return sb.String()
}

func collectText(msg protoreflect.Message, sb *strings.Builder) {
	msg.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		switch {
		case fd.IsMap():
			valFD := fd.MapValue()
			v.Map().Range(func(_ protoreflect.MapKey, mv protoreflect.Value) bool {
				appendLeaf(valFD, mv, sb)
				return true
			})
		case fd.IsList():
			list := v.List()
			for i := 0; i < list.Len(); i++ {
				appendLeaf(fd, list.Get(i), sb)
			}
		default:
			appendLeaf(fd, v, sb)
		}
		return true
	})
}

func appendLeaf(fd protoreflect.FieldDescriptor, v protoreflect.Value, sb *strings.Builder) {
	switch fd.Kind() {
	case protoreflect.MessageKind, protoreflect.GroupKind:
		collectText(v.Message(), sb)
	case protoreflect.StringKind:
		sb.WriteString(v.String())
		sb.WriteByte('\n')
	case protoreflect.BytesKind:
		sb.Write(v.Bytes())
		sb.WriteByte('\n')
	}
}
