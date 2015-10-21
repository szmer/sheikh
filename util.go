package main

import "encoding/json"

func toOdbRepr(thing interface{}) string {
	ret, _ := json.Marshal(thing)
	return string(ret)
}

type relSliceAggregate struct {
	masterIndex   *map[string][]vtxRel
	currentSlice  *[]vtxRel
	i             int
	currentMapKey string
}

func NewRelSliceAggregate(currentSlice []vtxRel, masterIndex map[string][]vtxRel) (rsa relSliceAggregate) {
	if currentSlice != nil {
		rsa.currentSlice = &currentSlice
	} else {
		*rsa.currentSlice = nil
	}
	rsa.masterIndex = &masterIndex
	return
}

func (rsa *relSliceAggregate) yield() (rel vtxRel) {
	if rsa.i < len(*rsa.currentSlice) {
		rsa.i++
		return (*rsa.currentSlice)[rsa.i-1]
	}
	if rsa.masterIndex == nil {
		return
	}
	var loadNextIndex bool
	if rsa.currentMapKey == "" { // first time
		loadNextIndex = true
	} else {
		loadNextIndex = false
	}
	for key, index := range *rsa.masterIndex {
		if loadNextIndex {
			rsa.currentSlice = &index
			rsa.currentMapKey = key
			return rsa.yield()
		}
		if key == rsa.currentMapKey {
			loadNextIndex = true
		}
	}
	return
}
