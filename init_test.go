// Copyright 2022 by Gilbert Ramirez <gram@alumni.rice.edu>

package objregexp

import (
	"testing"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner
func Test(t *testing.T) {
	// dlog.SetOutput(os.Stderr)
	TestingT(t)
}

type MySuite struct{}

var _ = Suite(&MySuite{})
