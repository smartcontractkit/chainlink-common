package codec

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"unicode"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

func NewRenamer(fields map[string]string) Modifier {
	m := &renamer{
		modifierBase: modifierBase[string]{
			fields:           fields,
			onToOffChainType: map[reflect.Type]reflect.Type{},
			offToOnChainType: map[reflect.Type]reflect.Type{},
		},
	}
	m.modifyFieldForInput = func(pkgPath string, field *reflect.StructField, _, newName string) error {
		field.Name = newName
		if unicode.IsLower(rune(field.Name[0])) {
			field.PkgPath = pkgPath
		}
		return nil
	}
	return m
}

type renamer struct {
	modifierBase[string]
}

func (r *renamer) TransformToOffChain(onChainValue any, itemType string) (any, error) {
	// itemType references the on-chain type
	// remap to the off-chain field name
	if itemType != "" {
		var ref string

		parts := strings.Split(itemType, ".")
		if len(parts) > 0 {
			ref = parts[len(parts)-1]
		}

		for on, off := range r.fields {
			if ref == on {
				// B.A -> C == B.C
				parts[len(parts)-1] = off
				itemType = strings.Join(parts, ".")

				break
			}
		}
	}

	rOutput, err := renameTransform(r.onToOffChainTyper, reflect.ValueOf(onChainValue), itemType)
	if err != nil {
		return nil, err
	}

	return rOutput.Interface(), nil
}

func (r *renamer) TransformToOnChain(offChainValue any, itemType string) (any, error) {
	log.Println(itemType)
	if itemType != "" {
		log.Println(itemType)
		var ref string

		parts := strings.Split(itemType, ".")
		if len(parts) > 0 {
			ref = parts[len(parts)-1]
		}

		for on, off := range r.fields {
			if ref == off {
				itemType = on

				break
			}
		}
	}

	rOutput, err := renameTransform(r.offToOnChainTyper, reflect.ValueOf(offChainValue), itemType)
	if err != nil {
		return nil, err
	}

	return rOutput.Interface(), nil
}

func renameTransform(
	typeFunc func(reflect.Type, string) (reflect.Type, error),
	rInput reflect.Value,
	itemType string,
) (reflect.Value, error) {
	toType, err := typeFunc(rInput.Type(), itemType)
	if err != nil {
		return reflect.Value{}, err
	}

	if toType == rInput.Type() {
		return rInput, nil
	}

	switch rInput.Kind() {
	case reflect.Pointer:
		return reflect.NewAt(toType.Elem(), rInput.UnsafePointer()), nil
	case reflect.Struct, reflect.Slice, reflect.Array:
		return transformNonPointer(toType, rInput)
	default:
		return reflect.Value{}, fmt.Errorf("%w: cannot rename kind %v", types.ErrInvalidType, rInput.Kind())
	}
}

func transformNonPointer(toType reflect.Type, rInput reflect.Value) (reflect.Value, error) {
	// make sure the input is addressable
	ptr := reflect.New(rInput.Type())
	reflect.Indirect(ptr).Set(rInput)

	// UnsafePointer is a bit of a Go hack but works because the data types/structure and data for the two types
	// are the same. The only change is the names of the fields.
	changed := reflect.NewAt(toType, ptr.UnsafePointer()).Elem()

	return changed, nil
}
