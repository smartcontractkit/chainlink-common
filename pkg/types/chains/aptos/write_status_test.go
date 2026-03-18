package aptos

import (
	"strings"
	"testing"
)

func TestClassifyWriteVmStatus(t *testing.T) {
	t.Run("receiver revert is terminal and marks receiver reverted", func(t *testing.T) {
		classification := ClassifyWriteVmStatus("Move abort in 0xabc::receiver::module: 42")
		if classification.Decision != WriteFailureDecisionTerminal {
			t.Fatalf("expected terminal, got %v", classification.Decision)
		}
		if classification.Kind != WriteFailureKindReceiverReverted {
			t.Fatalf("expected receiver reverted kind, got %v", classification.Kind)
		}
		if classification.ReceiverExecutionStatus != ReceiverExecutionStatusReverted {
			t.Fatalf("expected receiver reverted status, got %v", classification.ReceiverExecutionStatus)
		}
	})

	t.Run("known forwarder abort is terminal", func(t *testing.T) {
		classification := ClassifyWriteVmStatus("Move abort in 0x1::platform::forwarder: E_INVALID_SIGNATURE")
		if classification.Decision != WriteFailureDecisionTerminal {
			t.Fatalf("expected terminal, got %v", classification.Decision)
		}
		if classification.Kind != WriteFailureKindForwarderRejected {
			t.Fatalf("expected forwarder rejected kind, got %v", classification.Kind)
		}
	})

	t.Run("already processed is explicit decision", func(t *testing.T) {
		classification := ClassifyWriteVmStatus("Move abort in 0x1::platform::forwarder: E_ALREADY_PROCESSED")
		if classification.Decision != WriteFailureDecisionAlreadyProcessed {
			t.Fatalf("expected already processed, got %v", classification.Decision)
		}
		if classification.Kind != WriteFailureKindAlreadyProcessed {
			t.Fatalf("expected already processed kind, got %v", classification.Kind)
		}
	})

	t.Run("unknown forwarder abort remains retryable", func(t *testing.T) {
		classification := ClassifyWriteVmStatus("Move abort in 0x1::platform::forwarder: 99999")
		if classification.Decision != WriteFailureDecisionRetryable {
			t.Fatalf("expected retryable, got %v", classification.Decision)
		}
		if classification.Kind != WriteFailureKindUnknown {
			t.Fatalf("expected unknown kind, got %v", classification.Kind)
		}
	})

	t.Run("out of gas remains retryable", func(t *testing.T) {
		classification := ClassifyWriteVmStatus("Out of gas")
		if classification.Decision != WriteFailureDecisionRetryable {
			t.Fatalf("expected retryable, got %v", classification.Decision)
		}
		if classification.Message == "" {
			t.Fatal("expected out of gas message")
		}
	})

	t.Run("normalizeVmStatus strips simulated tx unexpected status prefix", func(t *testing.T) {
		// Prefix is stripped so the inner status is classified as forwarder terminal
		classification := ClassifyWriteVmStatus("simulated tx unexpected status: Move abort in 0x1::platform::forwarder: E_INVALID_SIGNATURE")
		if classification.Decision != WriteFailureDecisionTerminal {
			t.Fatalf("expected terminal after prefix strip, got %v", classification.Decision)
		}
		if classification.Kind != WriteFailureKindForwarderRejected {
			t.Fatalf("expected forwarder rejected, got %v", classification.Kind)
		}
	})

	t.Run("normalizeVmStatus strips simulate bad status prefix", func(t *testing.T) {
		classification := ClassifyWriteVmStatus("simulate bad status: Out of gas")
		if classification.Decision != WriteFailureDecisionRetryable {
			t.Fatalf("expected retryable after prefix strip, got %v", classification.Decision)
		}
		if classification.Reason != "transaction ran out of gas" {
			t.Fatalf("expected out of gas reason, got %q", classification.Reason)
		}
	})

	t.Run("forwarder location with ::forwarder suffix is forwarder", func(t *testing.T) {
		classification := ClassifyWriteVmStatus("Move abort in 0xaa::forwarder: E_INVALID_SIGNATURE")
		if classification.Decision != WriteFailureDecisionTerminal {
			t.Fatalf("expected terminal for ::forwarder location, got %v", classification.Decision)
		}
		if classification.Kind != WriteFailureKindForwarderRejected {
			t.Fatalf("expected forwarder rejected, got %v", classification.Kind)
		}
	})

	t.Run("known non-Move status transaction expired is retryable", func(t *testing.T) {
		classification := ClassifyWriteVmStatus("Transaction expired")
		if classification.Decision != WriteFailureDecisionRetryable {
			t.Fatalf("expected retryable, got %v", classification.Decision)
		}
		if classification.Reason != "transaction expired before inclusion" {
			t.Fatalf("expected expired reason, got %q", classification.Reason)
		}
	})

	t.Run("known non-Move status SEQUENCE_NUMBER_TOO_OLD is retryable", func(t *testing.T) {
		classification := ClassifyWriteVmStatus("SEQUENCE_NUMBER_TOO_OLD")
		if classification.Decision != WriteFailureDecisionRetryable {
			t.Fatalf("expected retryable, got %v", classification.Decision)
		}
		if !strings.Contains(strings.ToLower(classification.Reason), "sequence") {
			t.Fatalf("expected sequence in reason, got %q", classification.Reason)
		}
	})

	t.Run("known non-Move status Miscellaneous error is retryable", func(t *testing.T) {
		classification := ClassifyWriteVmStatus("Miscellaneous error")
		if classification.Decision != WriteFailureDecisionRetryable {
			t.Fatalf("expected retryable, got %v", classification.Decision)
		}
		if classification.Reason != "vm miscellaneous error" {
			t.Fatalf("expected miscellaneous reason, got %q", classification.Reason)
		}
	})

	t.Run("known non-Move status insufficient balance is retryable", func(t *testing.T) {
		classification := ClassifyWriteVmStatus("INSUFFICIENT_BALANCE_FOR_TRANSACTION_FEE")
		if classification.Decision != WriteFailureDecisionRetryable {
			t.Fatalf("expected retryable, got %v", classification.Decision)
		}
		if classification.Reason != "insufficient balance for transaction fee" {
			t.Fatalf("expected balance reason, got %q", classification.Reason)
		}
	})
}
