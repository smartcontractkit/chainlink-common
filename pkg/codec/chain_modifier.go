package codec

import (
	"reflect"

	"github.com/mitchellh/mapstructure"
)

func CreateModifier(config *ModificationConfig, hooks ...mapstructure.DecodeHookFunc) Modifier {
	return chainedModifier{hooks: hooks}
}

type chainedModifier struct {
	modifiers []Modifier
	hooks     []mapstructure.DecodeHookFunc
}

func (c chainedModifier) RetypeForInput(outputType reflect.Type) (reflect.Type, error) {
	//TODO implement me
	panic("implement me")
}

func (c chainedModifier) TransformInput(input any) (any, error) {
	//TODO implement me
	panic("implement me")
}

func (c chainedModifier) TransformOutput(output any) (any, error) {
	//TODO implement me
	panic("implement me")
}
