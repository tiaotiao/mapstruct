package mapstruct

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestMap2Struct(t *testing.T) {
	var args = struct {
		// basic
		Id    int64   `map:"id,required"`
		Name  string  `map:"name,required"`
		IsOK  bool    `map:"ok"`
		Price float64 `map:"price"`

		// options
		Ignore  string `map:"-"`
		NoName  string
		NoValue int64 `map:"novalue,1002"` // default value is "1002"
	}{}

	var vals = map[string]interface{}{
		"id":    1001,
		"name":  "tom",
		"ok":    true,
		"price": 29.9,

		"ignore":   "never mind", // discard
		"NoName":   "hello",
		"NotFound": "never mind", // discard
	}

	//  check
	check := args
	check.Id = 1001
	check.Name = "tom"
	check.IsOK = true
	check.Price = 29.9
	check.Ignore = ""
	check.NoName = "hello"
	check.NoValue = 1002

	checkMap2Struct(vals, &args, check, t)

	////////////////////////////////////////////////////////////

	var argsSlice = struct {

		// slice
		Tags   []string      `map:"tags"`
		Ids    []int64       `map:"ids"`
		Things []interface{} `map:"things"`
	}{}

	var valsSlice = map[string]interface{}{
		// string slice
		"tags": "edu,zhuhai,trees",
		"ids":  "10,20,30",
	}

	checkSlice := argsSlice
	checkSlice.Tags = []string{"edu", "zhuhai", "trees"}
	checkSlice.Ids = []int64{10, 20, 30}

	checkMap2Struct(valsSlice, &argsSlice, checkSlice, t)

	////////////////////////////////////////////////////////////

	var argsJson struct {

		// json
		Book      *BookInfo   `map:"book"`
		Books     []*BookInfo `map:"books"`
		BookNames []string    `map:"booknames"`
	}

	var valsJson = map[string]interface{}{
		// json
		"book":      json.RawMessage(`{"id":2001, "name":"harry poter"}`),
		"books":     json.RawMessage(`[{"id":3001, "name":"python programming"}, {"id":3002, "name":"cooking book"}]`),
		"booknames": json.RawMessage(`["c programming","thinking in java"]`),
	}

	checkJson := argsJson
	checkJson.Book = &BookInfo{2001, "harry poter"}
	checkJson.Books = []*BookInfo{{3001, "python programming"}, {3002, "cooking book"}}
	checkJson.BookNames = []string{"c programming", "thinking in java"}

	checkMap2Struct(valsJson, &argsJson, checkJson, t)
}

func checkMap2Struct(vals map[string]interface{}, args interface{}, check interface{}, t *testing.T) {
	err := Map2Struct(vals, args)
	if err != nil {
		t.Fatal(err.Error())
	}

	checkEqual(args, check, t)
}

func checkEqual(v1, v2 interface{}, t *testing.T) {
	rv1 := reflect.ValueOf(v1)
	rv2 := reflect.ValueOf(v2)
	if rv1.Kind() == reflect.Ptr {
		rv1 = rv1.Elem()
	}
	if rv2.Kind() == reflect.Ptr {
		rv2 = rv2.Elem()
	}

	for i := 0; i < rv1.NumField(); i++ {
		fv1 := rv1.Field(i).Interface()
		fv2 := rv2.Field(i).Interface()

		if !reflect.DeepEqual(fv1, fv2) {
			t.Fatal(fmt.Sprintf("not equal. %v: %#v!=%#v --> %v != %v \n", rv1.Type().Field(i).Name, fv1, fv2, v1, v2))
		}
	}
}

type BookInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
