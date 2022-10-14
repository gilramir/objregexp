// Copyright 2022 by Gilbert Ramirez <gram@alumni.rice.edu>

package objregexp

import "fmt"

type Compiler[T comparable] struct {
	finalized bool
	namespace map[string]ccType
	classMap  map[string]*Class[T]
	/*
		exemplarObj  map[string]T
		exemplarName map[T]string
	*/
}

type ccType int

const (
	ccClass = iota + 1
	ccExemplar
)

func NewCompiler[T comparable]() *Compiler[T] {
	s := &Compiler[T]{}
	s.Initialize()
	return s
}

func (s *Compiler[T]) Initialize() {
	s.assertNotFinalized()

	s.namespace = make(map[string]ccType)
	s.classMap = make(map[string]*Class[T])
	/*
		s.exemplarObj = make(map[string]T)
		s.exemplarName = make(map[T]string)
	*/
}

func (s *Compiler[T]) assertFinalized() {
	if !s.finalized {
		panic("objregexp.Compiler isn't finalized yet")
	}
}

func (s *Compiler[T]) assertNotFinalized() {
	if s.finalized {
		panic("objregexp.Compiler is already finalized")
	}
}

func (s *Compiler[T]) Finalize() {
	s.assertNotFinalized()
	s.finalized = true
}

func (s *Compiler[T]) RegisterClass(oClass *Class[T]) {

	if t, has := s.namespace[oClass.Name]; has {
		var msg string
		switch t {
		case ccClass:
			msg = fmt.Sprintf("A class with name '%s' already exists", oClass.Name)
		case ccExemplar:
			msg = fmt.Sprintf("An exemplar with name '%s' already exists", oClass.Name)
		}
		panic(msg)
	}
	s.classMap[oClass.Name] = oClass
	s.namespace[oClass.Name] = ccClass
}

func (s *Compiler[T]) Compile(text string) (*Regexp[T], error) {
	s.assertFinalized()

	factory := newNfaFactory[T](s)
	return factory.compile(text)
}
