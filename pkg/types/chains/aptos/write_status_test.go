package aptos

import "testing"

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
}
