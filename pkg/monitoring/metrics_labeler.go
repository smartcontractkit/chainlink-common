package monitoring

type MetricsLabeler struct {
	Labels map[string]string
}

func NewMetricsLabeler() MetricsLabeler {
	return MetricsLabeler{Labels: make(map[string]string)}
}

// With adds multiple key-value pairs to the CustomMessageLabeler for transmission With SendLogAsCustomMessage
func (c MetricsLabeler) With(keyValues ...string) MetricsLabeler {
	newCustomMetricsLabeler := NewMetricsLabeler()

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
