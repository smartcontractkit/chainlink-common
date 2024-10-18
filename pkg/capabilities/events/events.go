package events

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/beholder/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

const (
	// Duplicates the attributes in beholder/message.go::Metadata
	labelWorkflowOwner             = "workflow_owner_address"
	labelWorkflowID                = "workflow_id"
	labelWorkflowExecutionID       = "workflow_execution_id"
	labelWorkflowName              = "workflow_name"
	labelCapabilityContractAddress = "capability_contract_address"
	labelCapabilityID              = "capability_id"
	labelCapabilityVersion         = "capability_version"
	labelCapabilityName            = "capability_name"
)

type EmitMetadata struct {
	WorkflowOwner string // required
	WorkflowID    string // required
	WorkflowName  string // required

	WorkflowExecutionID       string // optional
	CapabilityContractAddress string // optional
	CapabilityID              string // optional
	CapabilityVersion         string // optional
	CapabilityName            string // optional
}

func (e EmitMetadata) merge(otherE EmitMetadata) EmitMetadata {
	owner := e.WorkflowOwner
	if otherE.WorkflowOwner != "" {
		owner = otherE.WorkflowOwner
	}

	wid := e.WorkflowID
	if otherE.WorkflowID != "" {
		wid = otherE.WorkflowID
	}

	eid := e.WorkflowExecutionID
	if otherE.WorkflowExecutionID != "" {
		eid = otherE.WorkflowExecutionID
	}

	name := e.WorkflowName
	if otherE.WorkflowName != "" {
		name = otherE.WorkflowName
	}

	addr := e.CapabilityContractAddress
	if otherE.CapabilityContractAddress != "" {
		addr = otherE.CapabilityContractAddress
	}

	capID := e.CapabilityID
	if otherE.CapabilityID != "" {
		capID = otherE.CapabilityID
	}

	capVersion := e.CapabilityVersion
	if otherE.CapabilityVersion != "" {
		capVersion = otherE.CapabilityVersion
	}

	capName := e.CapabilityName
	if otherE.CapabilityName != "" {
		capName = otherE.CapabilityName
	}

	return EmitMetadata{
		WorkflowOwner:             owner,
		WorkflowID:                wid,
		WorkflowExecutionID:       eid,
		WorkflowName:              name,
		CapabilityContractAddress: addr,
		CapabilityID:              capID,
		CapabilityVersion:         capVersion,
		CapabilityName:            capName,
	}
}

func (e EmitMetadata) attrs() []any {
	a := []any{}

	if e.WorkflowOwner != "" {
		a = append(a, labelWorkflowOwner, e.WorkflowOwner)
	}

	if e.WorkflowID != "" {
		a = append(a, labelWorkflowID, e.WorkflowID)
	}

	if e.WorkflowExecutionID != "" {
		a = append(a, labelWorkflowExecutionID, e.WorkflowExecutionID)
	}

	if e.WorkflowName != "" {
		a = append(a, labelWorkflowName, e.WorkflowName)
	}

	if e.CapabilityContractAddress != "" {
		a = append(a, labelCapabilityContractAddress, e.CapabilityContractAddress)
	}

	if e.CapabilityID != "" {
		a = append(a, labelCapabilityID, e.CapabilityID)
	}

	if e.CapabilityVersion != "" {
		a = append(a, labelCapabilityVersion, e.CapabilityVersion)
	}

	if e.CapabilityName != "" {
		a = append(a, labelCapabilityName, e.CapabilityName)
	}

	return a
}

type Emitter struct {
	client beholder.Emitter
	md     EmitMetadata
}

func (e *Emitter) With(md EmitMetadata) *Emitter {
	nmd := e.md.merge(md)
	return &Emitter{
		client: e.client,
		md:     nmd,
	}
}

func NewEmitter() *Emitter {
	return &Emitter{
		client: beholder.GetClient().Emitter,
	}
}

type Message struct {
	Msg      string
	Labels   map[string]any
	Metadata EmitMetadata
}

func (e *Emitter) Emit(ctx context.Context, msg Message) error {
	nmd := e.md.merge(msg.Metadata)

	if nmd.WorkflowOwner == "" {
		return errors.New("must provide workflow owner to emit event")
	}

	if nmd.WorkflowID == "" {
		return errors.New("must provide workflow id to emit event")
	}

	if nmd.WorkflowName == "" {
		return errors.New("must provide workflow name to emit event")
	}

	wm, err := values.WrapMap(msg.Labels)
	if err != nil {
		return fmt.Errorf("could not wrap map: %w", err)
	}

	pm := values.ProtoMap(wm)

	bytes, err := proto.Marshal(&pb.BaseMessage{
		Labels: pm,
		Msg:    msg.Msg,
	})
	if err != nil {
		return fmt.Errorf("could not marshal operational event: %w", err)
	}

	attrs := []any{
		"beholder_data_schema",
		"/capabilities-operational-event/versions/1",
		"beholder_data_type",
		"custom_message",
	}

	attrs = append(attrs, nmd.attrs()...)

	return e.client.Emit(
		ctx,
		bytes,
		attrs...,
	)
}
