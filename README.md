# Getting Start

This is a utility package of Golang provides convenient way to make conversion between maps and structs. There are two functions in this package:

> func **Map2Struct**(m maps[string]interface{}, s interface{}) (error)

> func **Struct2Map**(s interface{}) map[string]interface{}

Map2Struct() converts a map into a struct. You can custom how the fields of the struct match the keys in the map in a very simple way. You can also specify a default value or report an error if the key is not found. 

Struct2Map(), on the contrary, converts a struct into a map. This function is less powerful, honestly, but it's also helpful in some situations. 

To be clear, we use the term 'field' to indicate the field in struct. And we use term 'key' and 'value' to indicate key and value in map.

# Map2Struct

### Basic

This function converts a map into a struct. To use this function, import it first:

	import Map2Struct from mapstruct

Here is an example:

	m := map[string]interface{} {
		"Id"	: 1001,
		"Name"	: "tiaotiao",
	}
	
	s := struct {
		Id	int64		// 1001
		Name	string		// "tiaotiao"
	}{}
	
	err := Map2Struct(m, &s)

The type of map must be map[string]interface{}.

You can predefine the struct somewhere else, or just define a new type of struct where you need it. 

The second parameter must be a reference of the struct, so that it can be modified by the function.

Map2Struct() function will go through all the fields of the struct one by one, to look up if is there any key in the map matches the field. If there is a key in the map matches, the function will assign the value to the field.

By default, Map2Struct() use the exact name of the field to search for keys. It is case sensitive.

If the type of value is not the same as the type of field, Map2Struct() will try it best to convert and assign the value to the field. We will discuss it later.

If a field is not found, it will stay unchanged. But you are able to assign it a default value or report an error. We will discuss it later too.

Only public fields of the struct will be processed. It means **fields starting with lowercase will be ignored**.

### Name Tag

In most case, we want to custom another name for fields. For instance, if the field of struct is 'Name' but we want to match the key 'user_name' in the map. We can achieve this by give the field a tag.

Struct tag is a feature of Golang. You may probably see other well known struct tags like 'xml' or 'json' somewhere else. For more information please follow this [link](https://github.com/golang/go/wiki/Well-known-struct-tags).

	s := struct {
		Id	int64	`map:"user_id"`		// search for "user_id"
		Name	string	`map:"user_name"`	// search for "user_name"
		Website	string	`map:"-"`		// ignored
	}{}

We use 'map' to indicate this tag is for our package. The following string within the quote is the customed name for the field which you want to search for in the map.

Here is a special field tag "-" which indicate this field will be ignored by Map2Struct(). If a field tag is "-", this field will always stay unchanged after the function call, even if there is a key in the map matches this field exactly.

### Default Value

We are able to assign the field a default value if it's not found in the map. 

	s := struct {
		Id		int64	`map:"user_id,-1"`		// default value is -1
		Name		string	`map:"user_name,required"`	// must be found in map
		UserType	string	`map:"user_type"`		// no default value
		Blocked		bool	`map:"blocked,false"`		// default value is false
	}{}

Default value can be defined followed by the field tag, separated by a comma.

Default value only applies when field is not found.

Since default value is defined as type of string, the function will try it best to convert it to the target type.It will return an error if the conversion is failed.

If the default value is "required", Map2Struct() will return an error if it is not found.

### Advanced

	m := map[string]interface{}{
		"user_ids":		[]int64{501, 502, 503},
		"user_names":		`["batman", "ironman", "superman"]`,
		"group_info":		`{"id":1, "name":"heros"}`,
		"relative_groups":	json.RawMessage(`[{"id":7, "name":"FBI"}, {"id":9, "name":"CIA"}]`),
	}
	
	s := struct {
		Ids       []int64	`map:"user_ids"`	// [501, 502, 503]			support array
		Names     []string	`map:"user_names"`	// ["batman", "ironman", "superman"]	support json
		Group     group_info	`map:"group_info"`	// {1, "heros"}		support object from json string
		Relatives []*group_info	`map:"relative_groups"`	// [{7, "FBI"}, {9, "CIA"}]		support pointer
	}{}
	
	return Map2Struct(m, &s)

Type Conversion

Json

Object

Array

Pointer

# Struct2Map



