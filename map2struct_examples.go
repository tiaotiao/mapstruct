package mapstruct

import (
	"encoding/json"
)

func map2struct_example_1() error { // basic
	m := map[string]interface{}{
		"Id":   1001,
		"Name": "tiaotiao",
	}

	s := struct {
		Id   int64  // 1001
		Name string // "tiaotiao"
	}{}

	return Map2Struct(m, &s)
}

func map2struct_example_2() error { // not match
	m := map[string]interface{}{
		"Id":   1001,
		"name": "tiaotiao",
	}

	s := struct {
		ID   int64  // 0    case sensitive
		name string // ""   ignore fields starting with lower case
	}{}

	return Map2Struct(m, &s)
}

func map2struct_example_3() error { // using tag
	m := map[string]interface{}{
		"user_id":      1001,
		"user_name":    "tiaotiao",
		"user_website": "https://github.com/tiaotiao",
	}

	s := struct {
		Id      int64  `map:"user_id"`   // 1001         search for "user_id"
		Name    string `map:"user_name"` // "tiaotiao"   search for "user_name"
		Website string `map:"-"`         // ""           alway ignore "-"
	}{}

	return Map2Struct(m, &s)
}

func map2struct_example_4() error { // default value
	m := map[string]interface{}{
		"user_id":   1001,
		"user_name": "tiaotiao",
	}

	s := struct {
		Id   int64  `map:"user_id,-1"`         // 1001             will be -1 if not found
		Name string `map:"user_name,required"` // "tiaotiao"       will return an error if not found
		//                  "required" is the only reserved word which cannot be used as default value
		UserType string `map:"desc,normal"`   // "normal"         use default value
		Blocked  bool   `map:"blocked,false"` // false            use default value
	}{}

	return Map2Struct(m, &s)
}

type group_info struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

func map2struct_example_5() error { // advanced
	m := map[string]interface{}{
		"user_ids":        []int64{501, 502, 503},
		"user_names":      `["batman", "ironman", "superman"]`,
		"group_info":      `{"id":1, "name":"heros"}`,
		"relative_groups": json.RawMessage(`[{"id":7, "name":"FBI"}, {"id":9, "name":"CIA"}]`),
	}

	s := struct {
		Ids       []int64       // [501, 502, 503]                      support array
		Names     []string      // ["batman", "ironman", "superman"]    support json string
		Group     group_info    // {1, "heros"}                         support object from json string
		Relatives []*group_info // [{7, "FBI"}, {9, "CIA"}]         support pointer
	}{}

	return Map2Struct(m, &s)
}
