package bitflag

// Flag represents a single bitflag
type Flag uint

// Container represents a bitflag container and provides methods to simplify working with them
type Container uint

// EmptyContainer provides an empty bitflag container
const EmptyContainer Container = 0

// Has checks if the container has all the given flags set
func (cur Container) Has(flags ...Flag) bool {
	for _, flag := range flags {
		if uint(cur)&uint(flag) == 0 {
			return false
		}
	}
	return true
}

// With returns a new container with the given flags and the current ones set
func (cur Container) With(flags ...Flag) Container {
	val := uint(cur)
	for _, flag := range flags {
		val |= uint(flag)
	}
	return Container(val)
}

// Without returns a new container with the current flags but without the given ones set
func (cur Container) Without(flags ...Flag) Container {
	val := uint(cur)
	for _, flag := range flags {
		val &= ^uint(flag)
	}
	return Container(val)
}
