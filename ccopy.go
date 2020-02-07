// Package ccopy provides customizable deep copy of objects.
package ccopy

import (
	"errors"
	"fmt"
	"reflect"
)

const tagCcopy = "ccopy"

// Config represents the config for the customizable deep copy.
// Maps between tag value and functions that receive the tagged data and return the same data type.
type Config map[string]interface{}

// Copy deep copies an object respecting the customizations provided in the config.
// Unexported fields of a struct are ignored and will not be copied.
// The types unsafe.Pointer and uintptr are not supported and they will cause a panic.
// A channel will point to the original channel.
func (c Config) Copy(obj interface{}) (interface{}, error) {
	ov := reflect.ValueOf(obj)
	oc, err := c.copy(ov)
	if err != nil {
		return nil, err
	}
	return oc.Interface(), nil
}

func (c Config) copy(ov reflect.Value) (reflect.Value, error) {
	if !ov.IsValid() {
		return reflect.Value{}, errors.New("invalid value")
	}

	if t := ov.Type(); t.PkgPath() == "time" && t.Name() == "Time" {
		return c.copyTime(ov)
	}
	switch ov.Kind() {
	case reflect.Struct:
		return c.copyStruct(ov)
	case reflect.Ptr:
		return c.copyPointer(ov)
	case reflect.Slice:
		return c.copySlice(ov)
	case reflect.Map:
		return c.copyMap(ov)
	case reflect.Interface:
		return c.copyInterface(ov)
	case reflect.Array:
		return c.copyArray(ov)
	case reflect.Int, reflect.String, reflect.Int64, reflect.Float64, reflect.Bool, reflect.Uint, reflect.Uint64,
		reflect.Func, reflect.Chan, reflect.Float32,
		reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Complex64, reflect.Complex128,
		reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return ov, nil
	}
	panic(fmt.Sprintf("unsupported type: %s", ov.Kind()))
}

func (c Config) copyStruct(ov reflect.Value) (reflect.Value, error) {
	oc := reflect.New(ov.Type()).Elem()
	ot := ov.Type()
	for i := 0; i < ot.NumField(); i++ {
		// skip unexported fields
		if !ov.Field(i).CanInterface() {
			continue
		}
		tag := ot.Field(i).Tag.Get(tagCcopy)
		if tag == "" {
			// cannot set zero values, in case of pointers
			if v, err := c.copy(ov.Field(i)); err != nil {
				return reflect.Zero(ov.Type()), err
			} else if !v.IsZero() {
				oc.Field(i).Set(v)
			}
		} else {
			fn := c[tag]
			if fn == nil {
				return reflect.Zero(ov.Type()), fmt.Errorf("missing copy customiser for: %s", tag)
			}
			values := reflect.ValueOf(fn).Call([]reflect.Value{ov.Field(i)})
			// cannot set zero values, in case of pointers
			if !values[0].IsZero() {
				oc.Field(i).Set(values[0])
			}
		}
	}
	return oc, nil
}

func (c Config) copyPointer(ov reflect.Value) (reflect.Value, error) {
	if ov.IsNil() {
		return ov, nil
	}
	oc := reflect.New(ov.Type().Elem())
	v, err := c.copy(ov.Elem())
	if err != nil {
		return reflect.Zero(ov.Type()), err
	}
	if !v.IsZero() {
		oc.Elem().Set(v)
	}
	return oc, nil
}

func (c Config) copyInterface(ov reflect.Value) (reflect.Value, error) {
	if ov.IsNil() {
		return ov, nil
	}
	oc := reflect.New(ov.Type()).Elem()
	v, err := c.copy(ov.Elem())
	if err != nil {
		return reflect.Zero(ov.Type()), err
	}
	oc.Set(v)
	return oc, nil
}

func (c Config) copySlice(ov reflect.Value) (reflect.Value, error) {
	if ov.IsNil() {
		return ov, nil
	}
	oc := reflect.MakeSlice(ov.Type(), 0, ov.Len())
	for i := 0; i < ov.Len(); i++ {
		v, err := c.copy(ov.Index(i))
		if err != nil {
			return reflect.Zero(ov.Type()), err
		}
		oc = reflect.Append(oc, v)
	}
	return oc, nil
}

func (c Config) copyArray(ov reflect.Value) (reflect.Value, error) {
	oc := reflect.New(ov.Type()).Elem()
	slice := oc.Slice3(0, 0, ov.Len())
	for i := 0; i < ov.Len(); i++ {
		v, err := c.copy(ov.Index(i))
		if err != nil {
			return reflect.Zero(ov.Type()), err
		}
		slice = reflect.Append(slice, v)
	}
	return oc, nil
}

func (c Config) copyMap(ov reflect.Value) (reflect.Value, error) {
	if ov.IsNil() {
		return ov, nil
	}
	oc := reflect.MakeMapWithSize(ov.Type(), ov.Len())
	iter := ov.MapRange()
	for iter.Next() {
		k, err := c.copy(iter.Key())
		if err != nil {
			return reflect.Zero(ov.Type()), err
		}
		v, err := c.copy(iter.Value())
		if err != nil {
			return reflect.Zero(ov.Type()), err
		}
		oc.SetMapIndex(k, v)
	}
	return oc, nil
}

func (c Config) copyTime(ov reflect.Value) (reflect.Value, error) {
	return ov, nil
}
