package q

import (
	"fmt"
	"reflect"
)

type formatter struct {
	v     reflect.Value
	force bool
}

func (fo formatter) String() string {
	return fmt.Sprint(fo.v.Interface()) // unwrap it
}

func Sprint(a ...interface{}) string {
	return fmt.Sprint(wrap(a, true)...)
}

func wrap(a []interface{}, force bool) []interface{} {
	w := make([]interface{}, len(a))
	for i, x := range a {
		w[i] = formatter{v: reflect.ValueOf(x), force: force}
	}
	return w
}
