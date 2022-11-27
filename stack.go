package objregexp

// Modified
// from https://github.com/AlexandreChamard/go-generic

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

type stackT[T any] interface {
	Empty() bool
	Size() int
	Top() T
	Push(T)
	Pop()
}

func NewStack[T any]() stackT[T] { return &istackT[T]{} }

type istackT[T any] []T

func (this istackT[T]) Empty() bool  { return this.Size() == 0 }
func (this istackT[T]) Size() int    { return len(this) }
func (this istackT[T]) Top() T       { return this[len(this)-1] }
func (this istackT[T]) TopPtr() *T   { return &this[len(this)-1] }
func (this *istackT[T]) Push(info T) { *this = append(*this, info) }
func (this *istackT[T]) Pop()        { *this = (*this)[:len(*this)-1] }
