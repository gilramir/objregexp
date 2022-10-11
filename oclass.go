package objregexp

type Class[T comparable] struct {
	Name    string
	Matches func(T) bool
}
