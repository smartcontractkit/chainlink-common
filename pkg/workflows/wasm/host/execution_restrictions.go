package host

import (
	"strings"

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

type executionRestrictions struct {
	hasCaps       bool
	capType       sdk.CapabilityRestrictionType
	maxTotalCalls int32
	methods       map[methodKey]int32

	hasSecrets    bool
	maxSecrets    int32
	exactSecrets  map[secretKey]bool
	prefixSecrets []prefixRestriction
}

func newExecutionRestrictions(r *sdk.Restrictions) *executionRestrictions {
	er := &executionRestrictions{}
	if r == nil {
		return er
	}

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

	return er
}

func (e *executionRestrictions) callCapability(request *sdk.CapabilityRequest) bool {
	if e == nil || !e.hasCaps {
		return true
	}

	if e.maxTotalCalls == 0 {
		return false
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

func (e *executionRestrictions) getSecret(request *sdk.SecretRequest) bool {
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
