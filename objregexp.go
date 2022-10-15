// Copyright 2022 by Gilbert Ramirez <gram@alumni.rice.edu>

package objregexp

import (
	"fmt"
	"io/ioutil"
	"log"
)

// The debug logger for this module. By default, output is discarded.
var dlog *log.Logger

func init() {
	dlog = log.New(ioutil.Discard, "[DEBUG] ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
}

// Change the debug logger object in this module
func SetDebugLogger(logger *log.Logger) {
	dlog = logger
}

// Returns the output of the current debug logger in this module.
// You can then call SetOutput on the object, for example.
func GetDebugLoggerOutput() *log.Logger {
	return dlog
}

// The regex compiler. This holds all the user-defined classes
// that can be used in the regexes.
type Compiler[T any] struct {
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

// Instantiates and initializes a new Compiler.
func NewCompiler[T any]() *Compiler[T] {
	s := &Compiler[T]{}
	s.Initialize()
	return s
}

// Initializes a Compiler. Use this if you have allocated
// a Compiler object already, and only need to Initialize it.
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

// After registering all the classes, you must call Finalize() so
// Compile() can then be used.
func (s *Compiler[T]) Finalize() {
	s.assertNotFinalized()
	s.finalized = true
}

// Registers a user-defined class.
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

// Compile a regex string into a Regexp object.
// An error is returned if there is a syntax error.
func (s *Compiler[T]) Compile(text string) (*Regexp[T], error) {
	if !s.finalized {
		return nil, fmt.Errorf("The objregexp.Compiler is not finalized. Call Finalize().")
	}

	factory := newNfaFactory[T](s)
	return factory.compile(text)
}
