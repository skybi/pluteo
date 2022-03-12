package apikey

// Capability represents a single API key capability
type Capability uint

const (
	CapabilityReadMETARs Capability = 1 << iota
	CapabilityFeedMETARs
)

// Capabilities represents the container of API key capabilities.
// It provides methods Has, With and Without to check, set and unset certain capabilities.
type Capabilities uint

// EmptyCapabilities provides a capability container with no capabilities set
const EmptyCapabilities Capabilities = 0

// Has checks if the capability container has all the given capabilities set
func (cur Capabilities) Has(first Capability, others ...Capability) bool {
	if uint(cur)&uint(first) == 0 {
		return false
	}
	for _, other := range others {
		if uint(cur)&uint(other) == 0 {
			return false
		}
	}
	return true
}

// With returns a new capability container with all given and current capabilities set
func (cur Capabilities) With(first Capability, others ...Capability) Capabilities {
	val := uint(cur)
	val |= uint(first)
	for _, other := range others {
		val |= uint(other)
	}
	return Capabilities(val)
}

// Without returns a new capability container with the current and without the given capabilities set
func (cur Capabilities) Without(first Capability, others ...Capability) Capabilities {
	val := uint(cur)
	val &= ^uint(first)
	for _, other := range others {
		val &= ^uint(other)
	}
	return Capabilities(val)
}
