package mapstruct

import (
	"reflect"
	"strconv"
)

func Struct2Map(s interface{}) map[string]interface{} {
	return Struct2MapTag(s, DefaultTag)   // TODO use different tag?
}

func Struct2MapTag(s interface{}, tagName string) map[string]interface{} {
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

		name, option = parseTag(ft.Tag.Get(tagName))

		if name == "-" {
			continue // ignore "-"
		}

		if name == "" {
			name = ft.Name // use field name
		}

		if option == "omitempty" {
			if isEmpty(&fv) {
				continue // skip empty field
			}
		}

		// ft.Anonymous means embedded field
		if ft.Anonymous {
			if fv.Kind() == reflect.Ptr && fv.IsNil() {
				continue // nil
			}

			if (fv.Kind() == reflect.Struct) ||
				(fv.Kind() == reflect.Ptr && fv.Elem().Kind() == reflect.Struct) {

				// embedded struct
				embedded := Struct2MapTag(fv.Interface(), tagName)

				for embName, embValue := range embedded {
					m[embName] = embValue
				}
			}
			continue
		}

		if option == "string" {
			s := toString(fv)
			if s != nil {
				m[name] = s
				continue
			}
		}

		m[name] = fv.Interface()
	}

	return m
}

func toString(fv reflect.Value) interface{} {
	kind := fv.Kind()
	if kind == reflect.Int || kind == reflect.Int8 || kind == reflect.Int16 || kind == reflect.Int32 || kind == reflect.Int64 {
		return strconv.FormatInt(fv.Int(), 10)
	} else if kind == reflect.Uint || kind == reflect.Uint8 || kind == reflect.Uint16 || kind == reflect.Uint32 || kind == reflect.Uint64 {
		return strconv.FormatUint(fv.Uint(), 10)
	} else if kind == reflect.Float32 || kind == reflect.Float64 {
		return strconv.FormatFloat(fv.Float(), 'f', 2, 64)
	}
	// TODO support more types
	return nil
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
