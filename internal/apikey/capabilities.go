package apikey

import "github.com/skybi/data-server/internal/bitflag"

const (
	CapabilityReadMETARs bitflag.Flag = 1 << iota
	CapabilityFeedMETARs
)
