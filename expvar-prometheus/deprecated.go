package expvar_prometheus

// Deprecated: Use NewCollector. This name stutters.
func NewExpvarCollector() Collector {
	return NewCollector()
}
