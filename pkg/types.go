package pkg

import (
	"time"
)

const (
	requestMaxWaitTime = 5 * time.Second
	maxResponseLen     = 16 * 1024
)

const (
	MaxMetadataLen       = 1024 // https://cips.cardano.org/cips/cip6/
	RequestMaxWaitTime   = 1 * time.Second
	MaxWorkers           = 16
	WebsocketMethodName  = "Query"
	WebsocketServiceName = "p2p"
	WebsocketQueryType   = "jsonwsp/request"
	WebsocketVersion     = "1.0"
)
