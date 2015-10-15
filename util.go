package main

import "encoding/json"

func forceToString(thing interface{}) string {
	strThing, ok := thing.(string)
	if ok {
		return strThing
	}
	ret, _ := json.Marshal(thing)
	return string(ret)
}
