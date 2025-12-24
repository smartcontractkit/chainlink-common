package capabilities

import "time"

// TODO Implement BaseTriggerCapability - CRE-1523

type BaseTriggerCapability struct {
	/*
	 Keeps track of workflow registrations (similar to LLO streams trigger).
	 Handles retransmits based on T_retransmit and T_max.
	 Persists pending events in the DB to be resilient to node restarts.
	*/

	retransmitTime  time.Duration // time window for an event being ACKd before we retransmit
	maxEventTimeout time.Duration // timeout before events are considered lost if not ACKd
}

func (*BaseTriggerCapability) deliverEvent(event *TriggerEvent, workflowIDs []string) error {
	/*
	 Base Trigger Capability can interact with the Don2Don layer (in the remote capability setting)
	 as well as directly with a consumer (in the local setting).
	*/
	return nil // only when the event is successfully persisted and ready to be relaibly delivered
}
