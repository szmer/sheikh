package main

import (
	"chillson"
	"errors"
	"fmt"
)

type Vertex struct {
	Class, Rid     string
	Version        int
	diff           []string // changes since the last update to/from database
	propsContainer map[string]interface{}
	props          chillson.Son
}

func NewVertex(className string) (v Vertex) {
	v.Class = className
	v.propsContainer = make(map[string]interface{})
	v.props = chillson.Son{v.propsContainer}
	return
}

type Edge struct {
	Class, Name, Rid string
	From, To         *Vertex
}

func (v Vertex) Prop(name string) (interface{}, error) {
	return v.props.Get("[" + name + "]")
}

func (v Vertex) PropArr(name string) ([]interface{}, error) {
	return v.props.GetArr("[" + name + "]")
}

func (v Vertex) PropFloat(name string) (float64, error) {
	return v.props.GetFloat("[" + name + "]")
}

func (v Vertex) PropInt(name string) (int, error) {
	return v.props.GetInt("[" + name + "]")
}

func (v Vertex) PropObj(name string) (map[string]interface{}, error) {
	return v.props.GetObj("[" + name + "]")
}

// PropStr extracts vector's property as string (provided that it is defined for the Vector).
func (v Vertex) PropStr(name string) (string, error) {
	return v.props.GetStr("[" + name + "]")
}

/* SetProp takes a arbitrary number of property labels followed by their values. E.g.
SetProp("foo", "bar",  "baz", 5) assigns "bar" to "foo" property and 5 to "baz" property.
Method performs assignment in given order, and terminates if property label is not a string.
Arguments are not checked against schema constraints, which is left to the database. */
func (v *Vertex) SetProp(a ...interface{}) error {
	if len(a) == 0 && len(a)%2 != 0 {
		return errors.New("Vector SetProp: no arguments or odd number of arguments")
	}
	for i := 0; i < len(a); i += 2 {
		label, ok := a[i].(string)
		if !ok {
			return errors.New(fmt.Sprintf("Vector SetProp: non-string label %v", a[i]))
		}
		val, _ := a[i+1].(interface{})
		v.propsContainer[label] = val
		v.diff = append(v.diff, label)
	}
	return nil
}
