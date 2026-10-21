package visibility

// copyValue preserves unknown/known distinctions while preventing projected
// pointers from aliasing persistent observations or canonical state.
func copyValue[T any](value *T) *T {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

// Population zero is omitted by the existing wire schema; nil legacy observations
// therefore remain unavailable instead of being reconstructed from hidden cities.
func knownPopulation(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}
