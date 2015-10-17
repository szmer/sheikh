package main

import (
	"chillson"
)

type Vertex struct {
	Class, Name, Rid string
	props            chillson.Son
}

type Edge struct {
	Class, Name, Rid string
	From, To         *Vertex
}

func (v Vertex) ArrProp(name string) ([]interface{}, error) {
	return v.props.GetArr("[" + name + "]")
}

func (v Vertex) FloatProp(name string) (float64, error) {
	return v.props.GetFloat("[" + name + "]")
}

func (v Vertex) IntProp(name string) (int, error) {
	return v.props.GetInt("[" + name + "]")
}

func (v Vertex) ObjProp(name string) (map[string]interface{}, error) {
	return v.props.GetObj("[" + name + "]")
}

// StrProp extracts vector's property as string (provided that it is defined for the V-derived class).
func (v Vertex) StrProp(name string) (string, error) {
	return v.props.GetStr("[" + name + "]")
}
