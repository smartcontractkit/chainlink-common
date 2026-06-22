package host

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/actions/vault"
	caperrors "github.com/smartcontractkit/chainlink-common/pkg/capabilities/errors"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/actions/confidentialhttp"
	"github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

type methodKey struct {
	id     string
	method string
}

type secretKey struct {
	id        string
	namespace string
}

type prefixRestriction struct {
	prefix    string
	namespace string
	maxCalls  int32
}

// TODO refactor to instead be injected INTO the hepler
// this would allow raw secrets to call the same restriction check
// don't make it an execution helper itself
type executionRestrictions struct {
	ExecutionHelper
	mu sync.Mutex

	hasCaps       bool
	capType       sdk.CapabilityRestrictionType
	maxTotalCalls int32
	methods       map[methodKey]int32

	hasSecrets    bool
	maxSecrets    int32
	exactSecrets  map[secretKey]bool
	prefixSecrets []prefixRestriction
}

type executionRestrictionsWithRawSecrets struct {
	*executionRestrictions
}

func (e *executionRestrictionsWithRawSecrets) GetOwner() string {
	return e.ExecutionHelper.(ExecutionHelperWithRawSecrets).GetOwner()
}

func (e *executionRestrictionsWithRawSecrets) GetRawSecrets(ctx context.Context, request *sdk.GetSecretsRequest, fetcher EncryptionKeyFetcher) ([]*vault.SecretResponse, error) {
	rawSecretsHelper := e.ExecutionHelper.(ExecutionHelperWithRawSecrets)
	owner := rawSecretsHelper.GetOwner()

	e.mu.Lock()
	var allowed []*sdk.SecretRequest
	var responses []*vault.SecretResponse
	for _, req := range request.Requests {
		if e.reserveSecret(req) {
			allowed = append(allowed, req)
		} else {
			responses = append(responses, &vault.SecretResponse{
				Id: &vault.SecretIdentifier{
					Key:       req.Id,
					Namespace: req.Namespace,
					Owner:     owner,
				},
				Result: &vault.SecretResponse_Error{
					Error: fmt.Sprintf("secret %q in namespace %q denied by user pre-hook restrictions", req.Id, req.Namespace),
				},
			})
		}
	}
	e.mu.Unlock()

	if len(allowed) == 0 {
		return responses, nil
	}

	inner, err := rawSecretsHelper.GetRawSecrets(ctx, &sdk.GetSecretsRequest{Requests: allowed}, fetcher)
	if err != nil {
		return nil, err
	}
	return append(responses, inner...), nil
}

var _ ExecutionHelperWithRawSecrets = (*executionRestrictionsWithRawSecrets)(nil)

// NewRestrictedExecutionHelper wraps ExecutionHelper with restriction enforcement derived from r.
// If inner implements ExecutionHelperWithRawSecrets, the returned value will as well.
// If r is nil, ExecutionHelper is returned unchanged.
func NewRestrictedExecutionHelper(inner ExecutionHelper, r *sdk.Restrictions) ExecutionHelper {
	if r == nil {
		return inner
	}

	er := &executionRestrictions{ExecutionHelper: inner}

	if caps := r.Capabilities; caps != nil {
		er.hasCaps = true
		er.capType = caps.Type
		er.maxTotalCalls = caps.MaxTotalCalls
		er.methods = make(map[methodKey]int32)
		for _, cr := range caps.Restrictions {
			m, ok := cr.Restriction.(*sdk.CapabilityRestriction_Method)
			if !ok || m.Method == nil {
				continue
			}
			mr := m.Method
			key := methodKey{id: mr.Id, method: mr.Method}
			existing, found := er.methods[key]
			if !found || (mr.MaxCalls >= 0 && (existing < 0 || mr.MaxCalls < existing)) {
				er.methods[key] = mr.MaxCalls
			}
		}
	}

	if secrets := r.Secrets; secrets != nil {
		er.hasSecrets = true
		er.maxSecrets = secrets.MaxSecrets
		er.exactSecrets = make(map[secretKey]bool)
		for _, sr := range secrets.Restrictions {
			switch v := sr.Restriction.(type) {
			case *sdk.SecretRestriction_ExactSecret:
				s := v.ExactSecret
				er.exactSecrets[secretKey{id: s.Id, namespace: s.Namespace}] = true
			case *sdk.SecretRestriction_PrefixedSecret:
				p := v.PrefixedSecret
				er.prefixSecrets = append(er.prefixSecrets, prefixRestriction{
					prefix:    p.Prefix,
					namespace: p.Namespace,
					maxCalls:  p.MaxSecrets,
				})
			}
		}
	}

	if _, ok := inner.(ExecutionHelperWithRawSecrets); ok {
		return &executionRestrictionsWithRawSecrets{executionRestrictions: er}
	}
	return er
}

var confHttpRequest = (&confidentialhttp.ConfidentialHTTPRequest{}).ProtoReflect().Descriptor().FullName()

func (e *executionRestrictions) reserveCapabilityCall(request *sdk.CapabilityRequest) bool {
	if e == nil || !e.hasCaps {
		return true
	}

	if e.maxTotalCalls == 0 {
		return false
	}

	if request.Payload != nil {
		switch request.Payload.MessageName() {
		case confHttpRequest:
			conf := &confidentialhttp.ConfidentialHTTPRequest{}
			if err := request.Payload.UnmarshalTo(conf); err != nil {
				return false
			}

			secrets := conf.GetVaultDonSecrets()
			for _, secret := range secrets {
				if !e.reserveSecret(&sdk.SecretRequest{
					Id:        secret.Key,
					Namespace: secret.Namespace,
				}) {
					return false
				}
			}
		}
	}

	key := methodKey{id: request.Id, method: request.Method}
	remaining, found := e.methods[key]

	if !found {
		if e.capType == sdk.CapabilityRestrictionType_CAPABILITY_RESTRICTION_TYPE_CLOSED {
			return false
		}
		if e.maxTotalCalls > 0 {
			e.maxTotalCalls--
		}
		return true
	}

	if remaining == 0 {
		return false
	}

	if remaining > 0 {
		e.methods[key] = remaining - 1
	}
	if e.maxTotalCalls > 0 {
		e.maxTotalCalls--
	}
	return true
}

func (e *executionRestrictions) reserveSecret(request *sdk.SecretRequest) bool {
	if !e.hasSecrets {
		return true
	}

	if e.maxSecrets == 0 {
		return false
	}

	key := secretKey{id: request.Id, namespace: request.Namespace}
	exactMatch := e.exactSecrets[key]

	var matchedPrefixes []*prefixRestriction
	for i := range e.prefixSecrets {
		p := &e.prefixSecrets[i]
		if p.namespace == request.Namespace && strings.HasPrefix(request.Id, p.prefix) {
			if p.maxCalls == 0 {
				return false
			}
			matchedPrefixes = append(matchedPrefixes, p)
		}
	}

	if !exactMatch && len(matchedPrefixes) == 0 {
		return false
	}

	for _, p := range matchedPrefixes {
		if p.maxCalls > 0 {
			p.maxCalls--
		}
	}
	if e.maxSecrets > 0 {
		e.maxSecrets--
	}
	return true
}

func (e *executionRestrictions) CallCapability(ctx context.Context, request *sdk.CapabilityRequest) (*sdk.CapabilityResponse, error) {
	e.mu.Lock()
	allowed := e.reserveCapabilityCall(request)
	e.mu.Unlock()
	if !allowed {
		return nil, caperrors.NewLimitExceededError("call denied by user pre-hook restrictions", fmt.Errorf("%s %s", request.Id, request.Method))
	}
	return e.ExecutionHelper.CallCapability(ctx, request)
}

func (e *executionRestrictions) GetSecrets(ctx context.Context, request *sdk.GetSecretsRequest) ([]*sdk.SecretResponse, error) {
	e.mu.Lock()
	var allowed []*sdk.SecretRequest
	var responses []*sdk.SecretResponse
	for _, req := range request.Requests {
		if e.reserveSecret(req) {
			allowed = append(allowed, req)
		} else {
			responses = append(responses, &sdk.SecretResponse{
				Response: &sdk.SecretResponse_Error{
					Error: &sdk.SecretError{
						Id:        req.Id,
						Namespace: req.Namespace,
						Error:     fmt.Sprintf("secret %q in namespace %q denied by user pre-hook restrictions", req.Id, req.Namespace),
					},
				},
			})
		}
	}
	e.mu.Unlock()

	if len(allowed) == 0 {
		return responses, nil
	}

	inner, err := e.ExecutionHelper.GetSecrets(ctx, &sdk.GetSecretsRequest{Requests: allowed})
	if err != nil {
		return nil, err
	}
	return append(responses, inner...), nil
}
