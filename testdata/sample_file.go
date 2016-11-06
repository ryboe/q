package main

import "github.com/y0ssar1an/q"

func main() {
	a := 123
	b := "hello world"
	c := 3.1415926
	d := func(n int) bool { return n > 0 }(1)
	e := []int{1, 2, 3}
	f := []byte("goodbye world")
	g := e[1:]

	q.Q(a, b, c, d, e, f, g)
}
