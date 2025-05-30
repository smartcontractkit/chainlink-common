package beholder

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

func makeDynamicMessage(t *testing.T, pkg, msgName string) protoreflect.ProtoMessage {
	fdProto := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("test.proto"),
		Package: proto.String(pkg),
		MessageType: []*descriptorpb.DescriptorProto{{
			Name: proto.String(msgName),
		}},
	}

	fd, err := protodesc.NewFile(fdProto, nil)
	require.NoError(t, err)

	md := fd.Messages().ByName(protoreflect.Name(msgName))
	return dynamicpb.NewMessage(md)
}

func TestToSchemaPath(t *testing.T) {
	base := "/<base-path>"
	tests := []struct {
		pkg, msgName, expected string
	}{
		{
			pkg:      "alpha.bravo.charlie",
			msgName:  "FirstTest",
			expected: path.Join(base, "alpha/bravo/charlie/first_test.proto"),
		},
		{
			pkg:      "one.two",
			msgName:  "XMLEncode",
			expected: path.Join(base, "one/two/xmlencode.proto"),
		},
		{
			pkg:      "single",
			msgName:  "SimpleMessage",
			expected: path.Join(base, "single/simple_message.proto"),
		},
		{
			pkg:      "a.b.c.d.e",
			msgName:  "NestedLevel",
			expected: path.Join(base, "a/b/c/d/e/nested_level.proto"),
		},
		{
			pkg:     "mix.UpAndDOWN",
			msgName: "CamelCaseID",
			// package segment "UpAndDOWN" is left verbatim (no hyphenation), only the message gets snake_cased
			expected: path.Join(base, "mix/UpAndDOWN/camel_case_id.proto"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			m := makeDynamicMessage(t, tt.pkg, tt.msgName)
			got := toSchemaPath(m, base)
			assert.Equal(t, tt.expected, got)
		})
	}
}
