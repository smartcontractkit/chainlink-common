package aptos

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type ReceiverExecutionStatus uint8

const (
	ReceiverExecutionStatusUnknown ReceiverExecutionStatus = iota
	ReceiverExecutionStatusSuccess
	ReceiverExecutionStatusReverted
)

type WriteFailureDecision uint8

const (
	WriteFailureDecisionRetryable WriteFailureDecision = iota
	WriteFailureDecisionTerminal
	WriteFailureDecisionAlreadyProcessed
)

type WriteFailureKind uint8

const (
	WriteFailureKindUnknown WriteFailureKind = iota
	WriteFailureKindForwarderRejected
	WriteFailureKindReceiverReverted
	WriteFailureKindAlreadyProcessed
)

type WriteFailureClassification struct {
	Decision                WriteFailureDecision
	Kind                    WriteFailureKind
	ReceiverExecutionStatus ReceiverExecutionStatus
	Reason                  string
	Message                 string
}

func (c WriteFailureClassification) Retryable() bool {
	return c.Decision == WriteFailureDecisionRetryable
}

func (c WriteFailureClassification) AlreadyProcessed() bool {
	return c.Decision == WriteFailureDecisionAlreadyProcessed
}

func (c WriteFailureClassification) Terminal() bool {
	return c.Decision != WriteFailureDecisionRetryable
}

func (c WriteFailureClassification) MessagePtr() *string {
	if c.Message == "" {
		return nil
	}
	return &c.Message
}

var (
	moveAbortLocationRE = regexp.MustCompile(`(?i)move abort(?:ed)?(?: in)? ([^:]+(?:::[^:]+)+):?\s*(.*)$`)
	forwarderAbortNames = map[string]string{
		"E_INVALID_DATA_LENGTH":        "forwarder rejected the report because the report data was malformed",
		"E_INVALID_SIGNER":             "forwarder rejected the report because a signer was not part of the DON config",
		"E_DUPLICATE_SIGNER":           "forwarder rejected the report because the signer set contained duplicates",
		"E_INVALID_SIGNATURE_COUNT":    "forwarder rejected the report because the signature count was invalid",
		"E_INVALID_SIGNATURE":          "forwarder rejected the report because a signature was invalid",
		"E_ALREADY_PROCESSED":          "report was already processed by another node",
		"E_MALFORMED_SIGNATURE":        "forwarder rejected the report because a signature was malformed",
		"E_CALLBACK_DATA_NOT_CONSUMED": "forwarder callback data was not consumed by the receiver",
		"E_CONFIG_ID_NOT_FOUND":        "forwarder rejected the report because the DON config was not found",
		"E_INVALID_REPORT_VERSION":     "forwarder rejected the report because the report version was invalid",
	}
	forwarderAbortCodes = map[uint64]string{
		1:     "forwarder rejected the report because the report data was malformed",
		6:     "report was already processed by another node",
		12:    "forwarder callback data was not consumed by the receiver",
		15:    "forwarder rejected the report because the DON config was not found",
		16:    "forwarder rejected the report because the report version was invalid",
		65538: "forwarder rejected the report because a signer was not part of the DON config",
		65539: "forwarder rejected the report because the signer set contained duplicates",
		65540: "forwarder rejected the report because the signature count was invalid",
		65541: "forwarder rejected the report because a signature was invalid",
		65544: "forwarder rejected the report because a signature was malformed",
	}
)

func ClassifyWriteVmStatus(vmStatus string) WriteFailureClassification {
	vmStatus = normalizeVmStatus(vmStatus)
	if vmStatus == "" || strings.EqualFold(vmStatus, "Executed successfully") {
		return WriteFailureClassification{
			Decision: WriteFailureDecisionRetryable,
			Kind:     WriteFailureKindUnknown,
			Reason:   "no vm status available",
		}
	}

	if strings.EqualFold(vmStatus, "Out of gas") {
		return WriteFailureClassification{
			Decision: WriteFailureDecisionRetryable,
			Kind:     WriteFailureKindUnknown,
			Reason:   "transaction ran out of gas",
			Message:  vmStatus,
		}
	}

	// Explicit handling for known non-Move-abort Aptos VM statuses (retryable).
	if reason := knownNonMoveVmStatusReason(vmStatus); reason != "" {
		return WriteFailureClassification{
			Decision: WriteFailureDecisionRetryable,
			Kind:     WriteFailureKindUnknown,
			Reason:   reason,
			Message:  vmStatus,
		}
	}

	location, details, ok := splitMoveAbort(vmStatus)
	if !ok {
		return WriteFailureClassification{
			Decision: WriteFailureDecisionRetryable,
			Kind:     WriteFailureKindUnknown,
			Reason:   "vm failure was not a parsed move abort",
			Message:  vmStatus,
		}
	}

	if !isForwarderLocation(location) {
		return WriteFailureClassification{
			Decision:                WriteFailureDecisionTerminal,
			Kind:                    WriteFailureKindReceiverReverted,
			ReceiverExecutionStatus: ReceiverExecutionStatusReverted,
			Reason:                  "receiver or user module aborted",
			Message:                 fmt.Sprintf("receiver execution failed: %s", vmStatus),
		}
	}

	if name := extractAbortName(details); name != "" {
		if name == "E_ALREADY_PROCESSED" {
			return WriteFailureClassification{
				Decision: WriteFailureDecisionAlreadyProcessed,
				Kind:     WriteFailureKindAlreadyProcessed,
				Reason:   "forwarder reported the report was already processed",
				Message:  fmt.Sprintf("%s: %s", forwarderAbortNames[name], vmStatus),
			}
		}
		if msg, ok := forwarderAbortNames[name]; ok {
			return WriteFailureClassification{
				Decision: WriteFailureDecisionTerminal,
				Kind:     WriteFailureKindForwarderRejected,
				Reason:   "forwarder reported a terminal validation failure",
				Message:  fmt.Sprintf("%s: %s", msg, vmStatus),
			}
		}
	}

	if code, ok := extractAbortCode(details); ok {
		if code == 6 {
			return WriteFailureClassification{
				Decision: WriteFailureDecisionAlreadyProcessed,
				Kind:     WriteFailureKindAlreadyProcessed,
				Reason:   "forwarder reported the report was already processed",
				Message:  fmt.Sprintf("%s: %s", forwarderAbortCodes[code], vmStatus),
			}
		}
		if msg, found := forwarderAbortCodes[code]; found {
			return WriteFailureClassification{
				Decision: WriteFailureDecisionTerminal,
				Kind:     WriteFailureKindForwarderRejected,
				Reason:   "forwarder reported a terminal validation failure",
				Message:  fmt.Sprintf("%s: %s", msg, vmStatus),
			}
		}
	}

	return WriteFailureClassification{
		Decision: WriteFailureDecisionRetryable,
		Kind:     WriteFailureKindUnknown,
		Reason:   "forwarder abort was not a known terminal code",
		Message:  vmStatus,
	}
}

func normalizeVmStatus(vmStatus string) string {
	vmStatus = strings.TrimSpace(vmStatus)
	vmStatus = strings.TrimPrefix(vmStatus, "simulated tx unexpected status: ")
	vmStatus = strings.TrimPrefix(vmStatus, "simulate bad status: ")
	return strings.TrimSpace(vmStatus)
}

// knownNonMoveVmStatusReason returns a reason string for known non-Move-abort Aptos VM
// statuses (e.g. transaction expired, sequence errors). Returns "" if not a known status.
func knownNonMoveVmStatusReason(vmStatus string) string {
	lower := strings.ToLower(vmStatus)
	switch {
	case strings.Contains(lower, "transaction expired"):
		return "transaction expired before inclusion"
	case strings.Contains(lower, "sequence_number_too_old"), strings.Contains(lower, "sequence_number_too_new"):
		return "sequence number conflict; may need nonce resync"
	case strings.Contains(lower, "miscellaneous error"):
		return "vm miscellaneous error"
	case strings.Contains(lower, "insufficient_balance_for_transaction_fee"):
		return "insufficient balance for transaction fee"
	default:
		return ""
	}
}

func splitMoveAbort(vmStatus string) (location string, details string, ok bool) {
	matches := moveAbortLocationRE.FindStringSubmatch(vmStatus)
	if len(matches) == 3 {
		return strings.TrimSpace(matches[1]), strings.TrimSpace(matches[2]), true
	}
	if strings.Contains(strings.ToLower(vmStatus), "move abort") {
		return "", strings.TrimSpace(vmStatus), true
	}
	return "", "", false
}

func isForwarderLocation(location string) bool {
	location = strings.ToLower(strings.TrimSpace(location))
	return strings.HasSuffix(location, "::forwarder") || strings.Contains(location, "platform::forwarder")
}

func extractAbortName(details string) string {
	for _, token := range strings.FieldsFunc(details, func(r rune) bool {
		return !(r == '_' || ('A' <= r && r <= 'Z') || ('0' <= r && r <= '9'))
	}) {
		if strings.HasPrefix(token, "E_") {
			return token
		}
	}
	return ""
}

func extractAbortCode(details string) (uint64, bool) {
	for _, token := range strings.Fields(details) {
		token = strings.Trim(token, "(),.;")
		if strings.HasPrefix(strings.ToLower(token), "0x") {
			if value, err := strconv.ParseUint(token[2:], 16, 64); err == nil {
				return value, true
			}
		}
		if value, err := strconv.ParseUint(token, 10, 64); err == nil {
			return value, true
		}
	}
	return 0, false
}
