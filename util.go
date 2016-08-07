package sheikh

import "encoding/json"

// Convert a thing to OrientDB-syntax string representation.
func toOdbRepr(thing interface{}) string {
	ret, _ := json.Marshal(thing)
	return string(ret)
}
