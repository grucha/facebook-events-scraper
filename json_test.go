package fbevents

import (
	"testing"
)

func TestFindJSONInString_Object(t *testing.T) {
	html := `some text "myKey":{"a":"b","c":42} more text`
	res, err := findJSONInString(html, "myKey", nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.startIndex == -1 {
		t.Fatal("expected to find key, got -1")
	}
	m := asMap(res.data)
	if m == nil {
		t.Fatal("expected map, got nil")
	}
	if m["a"] != "b" {
		t.Errorf("expected a=b, got %v", m["a"])
	}
}

func TestFindJSONInString_Array(t *testing.T) {
	html := `"items":[1,2,3]`
	res, err := findJSONInString(html, "items", nil)
	if err != nil {
		t.Fatal(err)
	}
	s := asSlice(res.data)
	if len(s) != 3 {
		t.Errorf("expected 3 items, got %d", len(s))
	}
}

func TestFindJSONInString_Null(t *testing.T) {
	html := `"place": null`
	res, err := findJSONInString(html, "place", nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.startIndex == -1 {
		t.Fatal("expected to find key")
	}
	if res.data != nil {
		t.Errorf("expected nil data, got %v", res.data)
	}
}

func TestFindJSONInString_NotFound(t *testing.T) {
	html := `{"other":"value"}`
	res, err := findJSONInString(html, "missing", nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.startIndex != -1 {
		t.Errorf("expected -1, got %d", res.startIndex)
	}
}

func TestFindJSONInString_WithFilter(t *testing.T) {
	// Two occurrences; filter should pick the second
	html := `"data":{"x":1} ... "data":{"x":2,"target":true}`
	filter := func(v interface{}) bool {
		m := asMap(v)
		if m == nil {
			return false
		}
		_, ok := m["target"]
		return ok
	}
	res, err := findJSONInString(html, "data", filter)
	if err != nil {
		t.Fatal(err)
	}
	m := asMap(res.data)
	if m == nil {
		t.Fatal("expected map")
	}
	x, _ := m["x"].(float64)
	if x != 2 {
		t.Errorf("expected x=2, got %v", x)
	}
}

func TestFindJSONInString_EscapedQuotes(t *testing.T) {
	html := `"msg":{"text":"He said \"hello\" to her"}`
	res, err := findJSONInString(html, "msg", nil)
	if err != nil {
		t.Fatal(err)
	}
	m := asMap(res.data)
	if m == nil {
		t.Fatal("expected map")
	}
	text, _ := m["text"].(string)
	if text != `He said "hello" to her` {
		t.Errorf("unexpected text: %q", text)
	}
}

func TestFindJSONInString_NestedBraces(t *testing.T) {
	html := `"outer":{"inner":{"deep":"value"},"other":1}`
	res, err := findJSONInString(html, "outer", nil)
	if err != nil {
		t.Fatal(err)
	}
	m := asMap(res.data)
	if m == nil {
		t.Fatal("expected map")
	}
	inner := asMap(m["inner"])
	if inner == nil {
		t.Fatal("expected inner map")
	}
	if inner["deep"] != "value" {
		t.Errorf("expected deep=value, got %v", inner["deep"])
	}
}

func TestAsMap(t *testing.T) {
	var v interface{} = map[string]interface{}{"k": "v"}
	m := asMap(v)
	if m == nil || m["k"] != "v" {
		t.Errorf("asMap failed: %v", m)
	}

	if asMap("string") != nil {
		t.Error("asMap should return nil for non-map")
	}
}

func TestAsSlice(t *testing.T) {
	var v interface{} = []interface{}{1, 2, 3}
	s := asSlice(v)
	if len(s) != 3 {
		t.Errorf("expected 3, got %d", len(s))
	}

	if asSlice("string") != nil {
		t.Error("asSlice should return nil for non-slice")
	}
}

func TestGetString(t *testing.T) {
	m := map[string]interface{}{
		"a": map[string]interface{}{
			"b": "found",
		},
	}
	if getString(m, "a", "b") != "found" {
		t.Error("getString failed")
	}
	if getString(m, "a", "missing") != "" {
		t.Error("getString should return empty for missing key")
	}
	if getString(m, "missing", "b") != "" {
		t.Error("getString should return empty for missing nested key")
	}
}

func TestGetBool(t *testing.T) {
	m := map[string]interface{}{"flag": true}
	if !getBool(m, "flag") {
		t.Error("getBool failed")
	}
	if getBool(m, "missing") {
		t.Error("getBool should return false for missing key")
	}
}

func TestGetInt64(t *testing.T) {
	m := map[string]interface{}{"ts": float64(1681000200)}
	if getInt64(m, "ts") != 1681000200 {
		t.Errorf("getInt64 failed: %d", getInt64(m, "ts"))
	}
}
