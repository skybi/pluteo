package apikey

import "github.com/skybi/pluteo/internal/bitflag"

const (
	CapabilityReadMETARs bitflag.Flag = 1 << iota
	CapabilityFeedMETARs
)
