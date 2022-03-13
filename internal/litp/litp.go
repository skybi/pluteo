package litp

// String returns the pointer to a string literal
func String(val string) *string {
	return &val
}
