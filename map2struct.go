package mapstruct

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const DefaultTag = "map"

func Map2Struct(vals map[string]interface{}, dst interface{}) (err error) {
	return Map2StructTag(vals, dst, DefaultTag)
}

func Map2StructTag(vals map[string]interface{}, dst interface{}, tagName string) (err error) {
	defer func() {
		e := recover()
		if e != nil {
			if v, ok := e.(error); ok {
				err = fmt.Errorf("Panic: %v", v.Error())
			} else {
				err = fmt.Errorf("Panic: %v", e)
			}
		}
	}()

	pt := reflect.TypeOf(dst)
	pv := reflect.ValueOf(dst)

	if pv.Kind() != reflect.Ptr || pv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("not a pointer of struct")
	}

	var f reflect.StructField
	var ft reflect.Type
	var fv reflect.Value

	for i := 0; i < pt.Elem().NumField(); i++ {
		f = pt.Elem().Field(i)
		fv = pv.Elem().Field(i)
		ft = f.Type

		if f.Anonymous {
			continue
		}

		if !fv.CanSet() {
			continue
		}

		tag := f.Tag.Get(tagName)
		name, option := parseTag(tag)

		if name == "-" {
			continue // ignore "-"
		}

		if name == "" {
			// tag name is not set, use field name
			name = f.Name
		}

		// value from map
		val, ok := vals[name]

		if !ok { // value not found
			if option == "required" {
				return fmt.Errorf("'%v' is required", name)
			}

			if len(option) != 0 {
				val = option // 'option' means 'default value' here
			} else {
				continue // ignore it
			}
		}

		// assign or convert value to field
		if assignToField(val, name, fv) == nil {
			continue
		}

		switch v := val.(type) {
		case string:
			// parse string to value
			s := strings.TrimSpace(v)
			err = convertStringToValue(s, f.Name, fv, ft.Kind())

		case json.RawMessage:
			// unmarshal json
			err = convertJsonToValue(v, name, fv)

		default:

			err = fmt.Errorf("value type support: field=%v(%v) value=%v", f.Name, ft.Kind(), val)
		}

		if err != nil {
			return err
		}

		continue
	}

	return nil
}

func assignToField(val interface{}, name string, fv reflect.Value) error {
	vv := reflect.ValueOf(val)
	vt := reflect.TypeOf(val)
	ft := fv.Type()

	// assign or convert value to field
	if vt.AssignableTo(ft) {
		fv.Set(vv)
		return nil
	}
	if vt.ConvertibleTo(ft) {
		fv.Set(vv.Convert(ft))
		return nil
	}
	return fmt.Errorf("can not assign: %v(%v) value=%v(%v)", name, ft.Kind(), val, vt.Kind())
}

func convertJsonToValue(data json.RawMessage, name string, fv reflect.Value) error {
	var err error

	if fv.Kind() == reflect.Ptr {
		if fv.IsNil() {
			fv.Set(reflect.New(fv.Type().Elem()))
		}
	} else {
		fv = fv.Addr()
	}

	err = json.Unmarshal(data, fv.Interface())

	if err != nil {
		return fmt.Errorf("invalid json '%v': %v, %v", name, err.Error(), string(data))
	}

	return nil
}

func convertStringToValue(s string, name string, fv reflect.Value, kind reflect.Kind) error {
	if !fv.CanAddr() {
		return fmt.Errorf("can not addr: %v", name)
	}

	if assignToField(s, name, fv) == nil {
		return nil
	}

	if kind == reflect.String {
		fv.SetString(s)
		return nil
	}

	if kind == reflect.Slice {
		return convertStringToSlice(s, name, fv)
	}

	if kind == reflect.Ptr || kind == reflect.Struct {
		return convertJsonToValue(json.RawMessage(s), name, fv)
	}

	if kind == reflect.Bool {
		switch strings.ToLower(s) {
		case "true":
			fv.SetBool(true)
		case "false":
			fv.SetBool(false)
		case "1":
			fv.SetBool(true)
		case "0":
			fv.SetBool(false)
		default:
			return fmt.Errorf("invalid bool: %v value=%v", name, s)
		}
		return nil
	}

	if reflect.Int <= kind && kind <= reflect.Int64 {
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid int: %v value=%v", name, s)
		}
		fv.SetInt(i)

	} else if reflect.Uint <= kind && kind <= reflect.Uint64 {
		i, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid uint: %v value=%v", name, s)
		}
		fv.SetUint(i)

	} else if reflect.Float32 == kind || kind == reflect.Float64 {
		i, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return fmt.Errorf("invalid float: %v value=%v", name, s)
		}
		fv.SetFloat(i)

	} else {
		// type not support
		return fmt.Errorf("type not support: %v(%v) value=%v", name, kind.String(), s)
	}
	return nil
}

func convertStringToSlice(s string, name string, fv reflect.Value) error {
	var err error
	ft := fv.Type()
	et := ft.Elem()

	if len(s) == 0 {
		return nil
	}

	data := json.RawMessage(s)
	if data[0] == '[' && data[len(data)-1] == ']' {
		return convertJsonToValue(data, name, fv)
	}

	ss := strings.Split(s, ",")
	fs := reflect.MakeSlice(ft, 0, len(ss))

	for _, si := range ss {
		ev := reflect.New(et).Elem()

		err = convertStringToValue(si, name, ev, et.Kind())
		if err != nil {
			return err
		}
		fs = reflect.Append(fs, ev)
	}

	fv.Set(fs)

	return nil
}

func parseTag(tag string) (string, string) {
	tags := strings.Split(tag, ",")

	if len(tags) <= 0 {
		return "", ""
	}

	if len(tags) == 1 {
		return tags[0], ""
	}

	return tags[0], tags[1]
}
