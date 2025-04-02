package metrics

type Labeler struct {
	Labels map[string]string
}

func NewLabeler() Labeler {
	return Labeler{Labels: make(map[string]string)}
}

// With adds multiple key-value pairs to the Labeler to eventually be consumed by a Beholder metrics resource
func (c Labeler) With(keyValues ...string) Labeler {
	newCustomMetricsLabeler := NewLabeler()

	if len(keyValues)%2 != 0 {
		// If an odd number of key-value arguments is passed, return the original CustomMessageLabeler unchanged
		return c
	}

	// Copy existing labels from the current agent
	for k, v := range c.Labels {
		newCustomMetricsLabeler.Labels[k] = v
	}

	// Add new key-value pairs
	for i := 0; i < len(keyValues); i += 2 {
		key := keyValues[i]
		value := keyValues[i+1]
		newCustomMetricsLabeler.Labels[key] = value
	}

	return newCustomMetricsLabeler
}
