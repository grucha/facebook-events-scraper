package fbevents

import (
	"encoding/json"
	"fmt"
	"strings"
)

// jsonResult holds the result of a findJSONInString call.
type jsonResult struct {
	startIndex int
	endIndex   int
	data       interface{}
}

// findJSONInString searches haystack for a JSON value associated with key.
// The optional filter callback can reject candidates when the key appears
// multiple times in the HTML — pass nil to accept the first match.
// Returns a result with startIndex == -1 when the key is not found.
func findJSONInString(haystack, key string, filter func(interface{}) bool) (jsonResult, error) {
	searchKey := `"` + key + `":`
	offset := 0

	for {
		idx := strings.Index(haystack[offset:], searchKey)
		if idx == -1 {
			return jsonResult{startIndex: -1, endIndex: -1}, nil
		}

		// Absolute position of the start of the value (after `"key":`)
		valueStart := offset + idx + len(searchKey)

		// Skip optional whitespace between colon and value
		for valueStart < len(haystack) && (haystack[valueStart] == ' ' || haystack[valueStart] == '\t' || haystack[valueStart] == '\n' || haystack[valueStart] == '\r') {
			valueStart++
		}

		if valueStart >= len(haystack) {
			return jsonResult{startIndex: -1, endIndex: -1}, nil
		}

		startChar := haystack[valueStart]

		// Handle null literal
		if startChar == 'n' {
			if strings.HasPrefix(haystack[valueStart:], "null") {
				if filter == nil || filter(nil) {
					return jsonResult{startIndex: valueStart, endIndex: valueStart + 4, data: nil}, nil
				}
				offset = valueStart + 4
				continue
			}
		}

		if startChar != '{' && startChar != '[' {
			// Invalid start character — skip past this occurrence and keep searching
			offset = valueStart + 1
			continue
		}

		// Walk forward tracking brace/bracket depth and string state
		depth := 0
		inString := false
		end := -1

		for i := valueStart; i < len(haystack); i++ {
			c := haystack[i]
			if inString {
				if c == '\\' {
					i++ // skip escaped character
				} else if c == '"' {
					inString = false
				}
			} else {
				switch c {
				case '"':
					inString = true
				case '{', '[':
					depth++
				case '}', ']':
					depth--
					if depth == 0 {
						end = i
						goto foundEnd
					}
				}
			}
		}

	foundEnd:
		if end == -1 {
			offset = valueStart + 1
			continue
		}

		jsonStr := haystack[valueStart : end+1]
		var parsed interface{}
		if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
			// Malformed JSON at this occurrence; keep searching
			offset = end + 1
			continue
		}

		if filter != nil && !filter(parsed) {
			offset = end + 1
			continue
		}

		return jsonResult{startIndex: valueStart, endIndex: end, data: parsed}, nil
	}
}

// asMap safely type-asserts v to map[string]interface{}.
// Returns nil if v is not a JSON object.
func asMap(v interface{}) map[string]interface{} {
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	return nil
}

// asSlice safely type-asserts v to []interface{}.
// Returns nil if v is not a JSON array.
func asSlice(v interface{}) []interface{} {
	if s, ok := v.([]interface{}); ok {
		return s
	}
	return nil
}

// getString navigates a chain of keys through nested maps, returning the
// string value at the final key. Returns "" if any key is missing or if
// the final value is not a string.
func getString(m map[string]interface{}, keys ...string) string {
	var cur interface{} = m
	for _, k := range keys {
		cm := asMap(cur)
		if cm == nil {
			return ""
		}
		cur = cm[k]
		if cur == nil {
			return ""
		}
	}
	s, _ := cur.(string)
	return s
}

// getBool navigates nested maps and returns the bool value at the final key.
func getBool(m map[string]interface{}, keys ...string) bool {
	var cur interface{} = m
	for _, k := range keys {
		cm := asMap(cur)
		if cm == nil {
			return false
		}
		cur = cm[k]
		if cur == nil {
			return false
		}
	}
	b, _ := cur.(bool)
	return b
}

// getInt64 navigates nested maps and returns the int64 value at the final key.
// JSON numbers unmarshal as float64 into interface{}, so we convert.
func getInt64(m map[string]interface{}, keys ...string) int64 {
	var cur interface{} = m
	for _, k := range keys {
		cm := asMap(cur)
		if cm == nil {
			return 0
		}
		cur = cm[k]
		if cur == nil {
			return 0
		}
	}
	f, _ := cur.(float64)
	return int64(f)
}

// getFloat64 navigates nested maps and returns the float64 value at the final key.
func getFloat64(m map[string]interface{}, keys ...string) float64 {
	var cur interface{} = m
	for _, k := range keys {
		cm := asMap(cur)
		if cm == nil {
			return 0
		}
		cur = cm[k]
		if cur == nil {
			return 0
		}
	}
	f, _ := cur.(float64)
	return f
}

// hasKey returns true if the map contains the given key (value may be nil).
func hasKey(m map[string]interface{}, key string) bool {
	_, ok := m[key]
	return ok
}

// mustFindJSON is like findJSONInString but returns an error if the key is absent.
func mustFindJSON(haystack, key string, filter func(interface{}) bool, fieldName string) (jsonResult, error) {
	res, err := findJSONInString(haystack, key, filter)
	if err != nil {
		return jsonResult{}, fmt.Errorf("error parsing %s: %w", fieldName, err)
	}
	if res.startIndex == -1 {
		return jsonResult{}, fmt.Errorf("%s not found, please verify the event is publicly accessible", fieldName)
	}
	return res, nil
}
