package main

import (
	"chillson"
	"errors"
	"fmt"
)

func (c *Connection) registerClass(className string, place *map[string]regClassEntry) error {
	resp, err := (*c).Command(fmt.Sprintf("SELECT classes[name='%s'] FROM metadata:schema", className))
	if err != nil {
		return err
	}
	chill := chillson.Son{resp}
	sliceProps, err := chill.GetArr("[0][classes][properties]")
	if err != nil {
		return err
	}
	chill = chillson.Son{sliceProps}
	var entry regClassEntry
	entry.propList = "@RID"
	for i, _ := range sliceProps {
		var prop struct {
			name string
		}
		prop.name, err = chill.GetStr(fmt.Sprintf("[%v][name]", i))
		if err != nil {
			return errors.New(fmt.Sprintf("Unable to decode properties when trying to register a class %s: %v", className, resp))
		}
		entry.propList += ", " + prop.name
		entry.props = append(entry.props, prop)
	}
	(*place)[className] = entry
	return nil
}

func (c *Connection) RegisterEClass(name string) error {
	return (*c).registerClass(name, &(*c).regEClasses)
}

func (c *Connection) RegisterVClass(name string) error {
	return (*c).registerClass(name, &(*c).regVClasses)
}

func (c *Connection) SelectVertexes(class, cond, queryParams string, limit int) ([](*Vertex), error) {
	regEntry, ok := (*c).regVClasses[class]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Attempt to select vertexes of class %v, which is not registered", class))
	}
	comText := fmt.Sprintf("SELECT %s FROM %s%s%s LIMIT %v", regEntry.propList, class, " "+cond, " "+queryParams, limit)
	res, err := (*c).Command(comText)
	var ret [](*Vertex)
	for ind := range res {
		chill := chillson.Son{res[ind]}
		var v Vertex
		v.Class = class
		v.Name, err = chill.GetStr("[name]")
		if err == nil {
			v.Rid, err = chill.GetStr("[RID]")
		}
		if err != nil {
			return nil, err
		}
		props, _ := res[ind].(map[string]interface{})
		delete(props, "name")
		delete(props, "RID")
		for key := range props {
			if key[:1] == "@" {
				delete(props, key)
			}
		}
		v.props = chillson.Son{props}
		ret = append(ret, &v)
	}
	return ret, err
}
