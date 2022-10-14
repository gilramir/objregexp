// Copyright 2022 by Gilbert Ramirez <gram@alumni.rice.edu>

package objregexp

type Class[T any] struct {
	Name    string
	Matches func(T) bool
}
