package ocr2key

import (
	"bytes"
	"encoding/binary"
	"io"

	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/keystore/corekeys"
)

func UnmarshalMultichainPublicKey(d []byte) (map[string]ocrtypes.OnchainPublicKey, error) {
	m := map[string]ocrtypes.OnchainPublicKey{}
	buf := bytes.NewReader(d)

	for {
		// type
		typ, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}
		// length
		var length uint16
		err = binary.Read(buf, binary.LittleEndian, &length)
		if err != nil {
			return nil, err
		}
		// value
		pubKey := make([]byte, length)
		n, err := buf.Read(pubKey)
		if err != nil {
			return nil, err
		}
		if n != int(length) {
			return nil, io.EOF
		}

		k, err := corekeys.NewChainType(typ)
		if err != nil {
			// skipping unknown key type
			continue
		}
		m[string(k)] = pubKey

		if buf.Len() == 0 {
			break
		}
	}

	return m, nil
}
