package gorient

import "encoding/json"

func toOdbRepr(thing interface{}) string {
	ret, _ := json.Marshal(thing)
	return string(ret)
}

type relSliceAggregate struct {
	masterIndex   *map[string][]vtxRel
	currentSlice  *[]vtxRel
	keys          []string
	i             int
	currentMapKey string
}

func newRelSliceAggregate(currentSlice []vtxRel, masterIndex map[string][]vtxRel) (rsa relSliceAggregate) {
	if currentSlice != nil {
		rsa.currentSlice = &currentSlice
	} else {
		*rsa.currentSlice = nil
	}
	rsa.masterIndex = &masterIndex
	if masterIndex != nil {
		for key, _ := range masterIndex {
			rsa.keys = append(rsa.keys, key)
		}
	}
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
	for _, key := range rsa.keys {
		slice := (*rsa.masterIndex)[key]
		if loadNextIndex {
			rsa.currentSlice = &slice
			rsa.currentMapKey = key
			return rsa.yield()
		}
		if key == rsa.currentMapKey {
			loadNextIndex = true
		}
	}
	return
}
