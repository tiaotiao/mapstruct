package reflectutil

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const DefaultTag = "map"

func Map2Struct(vals map[string]interface{}, dst interface{}) (err error) {
	return Map2StructByTag(vals, dst, DefaultTag)
}

func Map2StructByTag(vals map[string]interface{}, dst interface{}, structTag string) (err error) {
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

		// parse struct tag
		tag := f.Tag.Get(structTag)
		name, option := parseTag(tag)
		if name == "-" {
			continue
		}
		if name == "" {
			name = strings.ToLower(f.Name)
		}

		// check tag option
		val, ok := vals[name]
		if !ok { // value not found
			if option == "required" {
				return fmt.Errorf("'%v' not found", name)
			}
			if len(option) != 0 {
				val = option // default value
			} else {
				//fv.Set(reflect.Zero(ft)) // TODO set zero value or just ignore it?
				continue
			}
		}

		// convert or set value to field
		vv := reflect.ValueOf(val)
		vt := reflect.TypeOf(val)

		if vt.Kind() != reflect.String {
			// try to assign and convert
			if vt.AssignableTo(ft) {
				fv.Set(vv)
				continue
			}

			if vt.ConvertibleTo(ft) {
				fv.Set(vv.Convert(ft))
				continue
			}

			return fmt.Errorf("value type not match: field=%v(%v) value=%v(%v)", f.Name, ft.Kind(), val, vt.Kind())
		}
		s := strings.TrimSpace(vv.String())
		if len(s) == 0 && option == "required" {
			return fmt.Errorf("value of required argument can't not be empty")
		}
		fk := ft.Kind()

		// convert string to value
		if fk == reflect.Ptr || fk == reflect.Struct {
			err = convertJsonValue(s, name, fv)
		} else if fk == reflect.Slice {
			err = convertSlice(s, f.Name, ft, fv)
		} else {
			err = convertValue(fk, s, f.Name, fv)
		}

		if err != nil {
			return err
		}
		continue
	}

	return nil
}

func convertSlice(s string, name string, ft reflect.Type, fv reflect.Value) error {
	var err error
	et := ft.Elem()

	if et.Kind() == reflect.Ptr || et.Kind() == reflect.Struct {
		return convertJsonValue(s, name, fv)
	}

	ss := strings.Split(s, ",")

	if len(s) == 0 || len(ss) == 0 {
		return nil
	}

	fs := reflect.MakeSlice(ft, 0, len(ss))

	for _, si := range ss {
		ev := reflect.New(et).Elem()

		err = convertValue(et.Kind(), si, name, ev)
		if err != nil {
			return err
		}
		fs = reflect.Append(fs, ev)
	}

	fv.Set(fs)

	return nil
}

func convertJsonValue(s string, name string, fv reflect.Value) error {
	var err error
	d := []byte(s)

	if fv.Kind() == reflect.Ptr {
		if fv.IsNil() {
			fv.Set(reflect.New(fv.Type().Elem()))
		}
	} else {
		fv = fv.Addr()
	}

	err = json.Unmarshal(d, fv.Interface())

	if err != nil {
		return fmt.Errorf("invalid json '%v': %v, %v", name, err.Error(), s)
	}

	return nil
}

func convertValue(kind reflect.Kind, s string, name string, fv reflect.Value) error {
	if !fv.CanAddr() {
		return fmt.Errorf("can not addr: %v", name)
	}

	if kind == reflect.String {
		fv.SetString(s)
		return nil
	}

	if kind == reflect.Bool {
		switch s {
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
			return fmt.Errorf("invalid int: %v value=%v", name, s)
		}
		fv.SetUint(i)

	} else if reflect.Float32 == kind || kind == reflect.Float64 {
		i, err := strconv.ParseFloat(s, 64)

		if err != nil {
			return fmt.Errorf("invalid float: %v value=%v", name, s)
		}

		fv.SetFloat(i)
	} else {
		// not support or just ignore it?
		// return fmt.Errorf("type not support: field=%v(%v) value=%v(%v)", name, ft.Kind(), val, vt.Kind())
	}
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

///////////////////////////////////////////////////////////////////////////////

func Struct2Map(s interface{}) map[string]interface{} {
	return Struct2MapByTag(s, DefaultTag)
}

func Struct2MapByTag(s interface{}, tagName string) map[string]interface{} {
	t := reflect.TypeOf(s)
	v := reflect.ValueOf(s)

	if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct {
		t = t.Elem()
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	m := make(map[string]interface{})

	for i := 0; i < t.NumField(); i++ {
		fv := v.Field(i)
		ft := t.Field(i)

		if !fv.CanInterface() {
			continue
		}

		if ft.PkgPath != "" { // unexported
			continue
		}

		var name string
		var option string
		tag := ft.Tag.Get(tagName)
		if tag != "" {
			ts := strings.Split(tag, ",")
			if len(ts) == 1 {
				name = ts[0]
			} else if len(ts) > 1 {
				name = ts[0]
				option = ts[1]
			}
			if name == "-" {
				continue // skip this field
			}
			if name == "" {
				name = strings.ToLower(ft.Name)
			}
			if option == "omitempty" {
				if isEmpty(&fv) {
					continue // skip empty field
				}
			}
		} else {
			name = strings.ToLower(ft.Name)
		}

		if ft.Anonymous && fv.Kind() == reflect.Ptr && fv.IsNil() {
			continue
		}
		if (ft.Anonymous && fv.Kind() == reflect.Struct) ||
			(ft.Anonymous && fv.Kind() == reflect.Ptr && fv.Elem().Kind() == reflect.Struct) {

			// embedded struct
			embedded := Struct2MapByTag(fv.Interface(), tagName)
			for embName, embValue := range embedded {
				m[embName] = embValue
			}
		} else if option == "string" {
			kind := fv.Kind()
			if kind == reflect.Int || kind == reflect.Int8 || kind == reflect.Int16 || kind == reflect.Int32 || kind == reflect.Int64 {
				m[name] = strconv.FormatInt(fv.Int(), 10)
			} else if kind == reflect.Uint || kind == reflect.Uint8 || kind == reflect.Uint16 || kind == reflect.Uint32 || kind == reflect.Uint64 {
				m[name] = strconv.FormatUint(fv.Uint(), 10)
			} else if kind == reflect.Float32 || kind == reflect.Float64 {
				m[name] = strconv.FormatFloat(fv.Float(), 'f', 2, 64)
			} else {
				m[name] = fv.Interface()
			}
		} else {
			m[name] = fv.Interface()
		}
	}

	return m
}

func isEmpty(v *reflect.Value) bool {
	k := v.Kind()
	if k == reflect.Bool {
		return v.Bool() == false
	} else if reflect.Int < k && k < reflect.Int64 {
		return v.Int() == 0
	} else if reflect.Uint < k && k < reflect.Uintptr {
		return v.Uint() == 0
	} else if k == reflect.Float32 || k == reflect.Float64 {
		return v.Float() == 0
	} else if k == reflect.Array || k == reflect.Map || k == reflect.Slice || k == reflect.String {
		return v.Len() == 0
	} else if k == reflect.Interface || k == reflect.Ptr {
		return v.IsNil()
	}
	return false
}
