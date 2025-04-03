package write_target

import (
	"fmt"

	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

// AsError returns the WriteError message as an (Go) error
func (e *WriteError) AsError() error {
	protoName := protoimpl.X.MessageTypeOf(e).Descriptor().FullName()
	return fmt.Errorf("%s [ERR-%v] - %s: %s", protoName, e.Code, e.Summary, e.Cause)
}
