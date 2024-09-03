package main

import (
	_ "github.com/plentymarkets/mc-telemetry-driver/pkg/teldrvr"
	"github.com/plentymarkets/mc-telemetry/pkg/telemetry"
)

func main() {
	telemetry.SetDriver("nrZerolog")
	telemetry.SetTraceDriver("nrZerolog")

	transaction, err := telemetry.Start("test transaction")
	if err != nil {
		panic(err)
	}

	sID := transaction.SegmentStart("test segment")
	transaction.SegmentEnd(sID)
	transaction.Done()
}
