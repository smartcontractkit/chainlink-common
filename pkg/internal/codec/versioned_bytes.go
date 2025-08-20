package codec

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/fxamacker/cbor/v2"
	jsonv2 "github.com/go-json-experiment/json"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
	valuespb "github.com/smartcontractkit/chainlink-protos/cre/go/values/pb"
)

type EncodingVersion uint32

func (v EncodingVersion) Uint32() uint32 {
	return uint32(v)
}

// enum of all known encoding formats for versioned data.
const (
	JSONEncodingVersion1 EncodingVersion = iota
	JSONEncodingVersion2
	CBOREncodingVersion
	ValuesEncodingVersion
)

const DefaultEncodingVersion = CBOREncodingVersion

func EncodeVersionedBytes(data any, version EncodingVersion) (*VersionedBytes, error) {
	var byt []byte
	var err error

	switch version {
	case JSONEncodingVersion1:
		byt, err = json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", types.ErrInvalidType, err)
		}
	case JSONEncodingVersion2:
		byt, err = jsonv2.Marshal(data, jsonv2.StringifyNumbers(true))
		if err != nil {
			return nil, fmt.Errorf("%w: %w", types.ErrInvalidType, err)
		}
	case CBOREncodingVersion:
		enco := cbor.CoreDetEncOptions()
		enco.Time = cbor.TimeRFC3339Nano
		var enc cbor.EncMode
		enc, err = enco.EncMode()
		if err != nil {
			return nil, fmt.Errorf("%w: %w", types.ErrInternal, err)
		}
		byt, err = enc.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", types.ErrInvalidType, err)
		}
	case ValuesEncodingVersion:
		val, err := values.Wrap(data)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", types.ErrInvalidType, err)
		}
		byt, err = proto.Marshal(values.Proto(val))
		if err != nil {
			return nil, fmt.Errorf("%w: %w", types.ErrInvalidType, err)
		}
	default:
		return nil, fmt.Errorf("%w: unsupported encoding version %d for data %v", types.ErrInvalidEncoding, version, data)
	}

	return &VersionedBytes{Version: version.Uint32(), Data: byt}, nil
}

func DecodeVersionedBytes(res any, vData *VersionedBytes) error {
	if vData == nil {
		return errors.New("cannot decode nil versioned bytes")
	}

	var err error
	switch EncodingVersion(vData.Version) {
	case JSONEncodingVersion1:
		decoder := json.NewDecoder(bytes.NewBuffer(vData.Data))
		decoder.UseNumber()

		err = decoder.Decode(res)
	case JSONEncodingVersion2:
		err = jsonv2.Unmarshal(vData.Data, res, jsonv2.StringifyNumbers(true))
	case CBOREncodingVersion:
		decopt := cbor.DecOptions{UTF8: cbor.UTF8DecodeInvalid}
		var dec cbor.DecMode
		dec, err = decopt.DecMode()
		if err != nil {
			return fmt.Errorf("%w: %w", types.ErrInternal, err)
		}
		err = dec.Unmarshal(vData.Data, res)
	case ValuesEncodingVersion:
		protoValue := &valuespb.Value{}
		err = proto.Unmarshal(vData.Data, protoValue)
		if err != nil {
			return fmt.Errorf("%w: %w", types.ErrInvalidType, err)
		}

		var value values.Value
		value, err = values.FromProto(protoValue)
		if err != nil {
			return fmt.Errorf("%w: %w", types.ErrInvalidType, err)
		}

		valuePtr, ok := res.(*values.Value)
		if ok {
			*valuePtr = value
		} else {
			err = value.UnwrapTo(res)
			if err != nil {
				return fmt.Errorf("%w: %w", types.ErrInvalidType, err)
			}
		}
	default:
		return fmt.Errorf("unsupported encoding version %d for versionedData %v", vData.Version, vData.Data)
	}

	if err != nil {
		return fmt.Errorf("%w: %w", types.ErrInvalidType, err)
	}

	return nil
}
