package main

import "encoding/json"

func toOdbRepr(thing interface{}) string {
	ret, _ := json.Marshal(thing)
	return string(ret)
}
