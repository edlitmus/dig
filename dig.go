/*
  Copyright (c) 2013 José Carlos Nieto, http://xiam.menteslibres.org/

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

package dig

import (
	"fmt"
	"reflect"

	"menteslibres.net/gosexy/to"
)

// Bool Returns a boolean starting from a Slice or Map.
func Bool(src interface{}, route ...interface{}) bool {
	var b bool
	err := Get(src, &b, route...)
	if err != nil {
		return false
	}
	return b
}

// Uint64 Returns an uint64 starting from a Slice or Map.
func Uint64(src interface{}, route ...interface{}) uint64 {
	var i uint64
	err := Get(src, &i, route...)
	if err != nil {
		return uint64(0)
	}
	return i
}

// Int64 Returns an int64 starting from a Slice or Map.
func Int64(src interface{}, route ...interface{}) int64 {
	var i int64
	err := Get(src, &i, route...)
	if err != nil {
		return int64(0)
	}
	return i
}

// Float32 Returns a float32 starting from a Slice or Map.
func Float32(src interface{}, route ...interface{}) float32 {
	var f float32
	err := Get(src, &f, route...)
	if err != nil {
		return float32(0)
	}
	return f
}

// Float64 Returns a float64 starting from a Slice or Map.
func Float64(src interface{}, route ...interface{}) float64 {
	var f float64
	err := Get(src, &f, route...)
	if err != nil {
		return float64(0)
	}
	return f
}

// Interface Returns an interface{} starting from a Slice or Map.
func Interface(src interface{}, route ...interface{}) interface{} {
	var i interface{}
	err := Get(src, &i, route...)
	if err != nil {
		return nil
	}
	return i
}

// String Returns a string starting from a Slice or Map.
func String(src interface{}, route ...interface{}) string {
	var s string
	err := Get(src, &s, route...)
	if err != nil {
		return ""
	}
	return s
}

// pick Returns the element of the Slice or Map given by route.
func pick(src interface{}, dig bool, route ...interface{}) (*reflect.Value, error) {
	var err error

	v := reflect.ValueOf(src)

	if v.Kind() != reflect.Ptr || v.IsNil() {
		return nil, fmt.Errorf("source is not a pointer")
	}

	v = v.Elem()

	for _, key := range route {
		u := v
		switch kind := v.Kind(); kind {
		case reflect.Slice:
			switch i := key.(type) {
			case int:
				if i < v.Len() {
					v = v.Index(i)
				} else {
					return nil, fmt.Errorf("undefined index: %d for kind: %s", i, kind)
				}
			}
		case reflect.Map:
			vkey := reflect.ValueOf(key)
			v = v.MapIndex(vkey)
			if dig && !v.IsValid() {
				u.SetMapIndex(vkey, reflect.MakeMap(u.Type()))
				v = u.MapIndex(vkey)
			}
			if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
				v = v.Elem()
			}
		}
		if v.IsValid() {
			if v.CanInterface() {
				v = reflect.ValueOf(v.Interface())
			}
		}
	}

	return &v, err
}

// Set Starts with src (pointer to Slice or Map) tries to follow the given route, if
// the route is found it then tries to set the node with the given value (val).
func Set(src interface{}, val interface{}, route ...interface{}) error {
	l := len(route)

	if l < 1 {
		return fmt.Errorf("missing route")
	}

	parent := route[0 : l-1]
	last := route[l-1 : l]

	p, err := pick(src, false, parent...)

	if err != nil {
		return err
	}

	if !p.IsValid() {
		return fmt.Errorf("the given route does not exists")
	}

	p.SetMapIndex(reflect.ValueOf(last[0]), reflect.ValueOf(val))

	return nil
}

// Get Starts with src (pointer to Slice or Map) tries to follow the given route,
// if the route is found it then tries to copy or convert the found node into
// the value pointed by dst.
func Get(src interface{}, dst interface{}, route ...interface{}) error {

	if len(route) < 1 {
		return fmt.Errorf("missing route")
	}

	dv := reflect.ValueOf(dst)

	if dv.Kind() != reflect.Ptr || dv.IsNil() {
		return fmt.Errorf("destination is not a pointer")
	}

	sv := reflect.ValueOf(src)

	if sv.Kind() != reflect.Ptr || sv.IsNil() {
		return fmt.Errorf("source is not a pointer")
	}

	// Setting to zero before setting it again.
	dv.Elem().Set(reflect.Zero(dv.Elem().Type()))

	p, err := pick(src, false, route...)

	if err != nil {
		return err
	}

	if !p.IsValid() {
		return fmt.Errorf("could not find the route: %#v", route)
	}

	if dv.Elem().Type() != p.Type() {
		// Trying conversion
		if p.CanInterface() {
			t, err := to.Convert(p.Interface(), dv.Elem().Kind())
			if err == nil {
				tv := reflect.ValueOf(t)
				if dv.Elem().Type() == tv.Type() {
					p = &tv
				}
			}
		}
	}

	if dv.Elem().Type() == p.Type() || dv.Elem().Kind() == reflect.Interface {
		dv.Elem().Set(*p)
	} else {
		return fmt.Errorf("could not assign %s to %s", p.Type(), dv.Elem().Type())
	}

	return nil
}

// Dig Makes a path to the given route, if the route already exists it overwrites it
// with a zero value.
func Dig(src interface{}, route ...interface{}) error {
	v, err := pick(src, true, route...)
	if !v.IsValid() {
		return fmt.Errorf("could not reach node")
	}
	return err
}
