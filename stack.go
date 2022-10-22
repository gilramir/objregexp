package objregexp

// From https://github.com/AlexandreChamard/go-generic

/*
Copyright (c) 2022 Alexandre Chamard-Bois

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
"Software"), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

type Stack[T any] interface {
	Empty() bool
	Size() int
	Top() T
	Push(T)
	Pop()
}

func NewStack[T any]() Stack[T] { return &stackT[T]{} }

type stackT[T any] []T

func (this stackT[T]) Empty() bool  { return this.Size() == 0 }
func (this stackT[T]) Size() int    { return len(this) }
func (this stackT[T]) Top() T       { return this[len(this)-1] }
func (this stackT[T]) TopPtr() *T   { return &this[len(this)-1] }
func (this *stackT[T]) Push(info T) { *this = append(*this, info) }
func (this *stackT[T]) Pop()        { *this = (*this)[:len(*this)-1] }
