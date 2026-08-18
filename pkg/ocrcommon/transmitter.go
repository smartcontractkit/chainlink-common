package ocrcommon

import (
	"context"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

// TransmitterWithAccount answers FromAccount with the account an oracle is
// registered under, leaving the rest of the transmitter alone.
//
// The account is part of the identity a configuration lists, alongside the peer ID
// and the public keys, and libocr checks all of them: an oracle whose account does
// not match its entry is not recognised as a member at all.
//
// It is supplied from outside rather than reached for from within because it
// belongs to the node, not to whatever the transmitter transmits. A plugin's own
// transmitter answers for where a report goes - which for a capability is usually
// back to whoever asked - and who it is transmitting as is not its to know.
func TransmitterWithAccount(account ocrtypes.Account, transmitter ocr3types.ContractTransmitter[[]byte]) ocr3types.ContractTransmitter[[]byte] {
	return transmitterWithAccount{ContractTransmitter: transmitter, account: account}
}

type transmitterWithAccount struct {
	ocr3types.ContractTransmitter[[]byte]
	account ocrtypes.Account
}

func (t transmitterWithAccount) FromAccount(context.Context) (ocrtypes.Account, error) {
	return t.account, nil
}
