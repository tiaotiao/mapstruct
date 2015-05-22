package mapstruct

import (
	"fmt"
	"reflect"
	"testing"
)

func TestStruct2Map(t *testing.T) {
	var args = struct {
		Id        int64  `map:"id"`
		Name      string `map:"name"`
		IsOK      bool   `map:"ok"`
		OmitEmpty string `map:"empty,omitempty"` // discard
		Ignore    string `map:"-"`               // discard
		NoName    string
		StringInt int64 `map:"strint,string"`
	}{
		Id:        1001,
		Name:      "tim",
		IsOK:      true,
		OmitEmpty: "",
		Ignore:    "never mind",
		NoName:    "hello",
		StringInt: 2001,
	}

	var check = map[string]interface{}{
		"id":     int64(1001),
		"name":   "tim",
		"ok":     true,
		"NoName": "hello",
		"strint": "2001",
	}

	vals := Struct2Map(args)

	checkStruct2Map(vals, check, t)
}

func checkStruct2Map(vals, check map[string]interface{}, t *testing.T) {
	for k, v := range check {
		cv, ok := vals[k]
		if !ok {
			t.Fatal(fmt.Sprintf("key [%v] not found, %v", k, v))
		}

		if !reflect.DeepEqual(v, cv) {
			t.Fatal(fmt.Sprintf("not equal [%v]: %v != %v", k, v, cv))
		}

		delete(check, k)
		delete(vals, k)
	}

	if len(vals) > 0 {
		t.Fatal(fmt.Sprintf("unexpected vals [%v]", vals))
	}
}
