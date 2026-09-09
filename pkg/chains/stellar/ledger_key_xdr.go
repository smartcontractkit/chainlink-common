package stellar

import (
	"bytes"
	"fmt"

	"github.com/stellar/go-stellar-sdk/xdr"
)

func validateLedgerKeyXDR(b []byte) error {
	if len(b) == 0 {
		return fmt.Errorf("empty XDR")
	}

	var key xdr.LedgerKey
	if err := key.UnmarshalBinary(b); err != nil {
		return err
	}

	encoded, err := key.MarshalBinary()
	if err != nil {
		return err
	}
	if !bytes.Equal(encoded, b) {
		return fmt.Errorf("trailing %d bytes", len(b)-len(encoded))
	}
	return nil
}
