package main

import (
	"chillson"
	"errors"
	"fmt"
	"strings"
)

func (c *Connection) registerClass(name string, place *map[string]regClassEntry) error {
	resp, err := (*c).Command(fmt.Sprintf("SELECT classes[name='%s'] FROM metadata:schema", name))
	if err != nil {
		return err
	}
	chill := chillson.Son{resp}
	sliceProps, err := chill.GetArr("[0][classes][properties]")
	if err != nil {
		return err
	}
	chill = chillson.Son{sliceProps}
	fmt.Printf("%v\n", sliceProps)
	var propNames []string
	for i, _ := range sliceProps {
		prop, err := chill.GetStr(fmt.Sprintf("[%v][name]", i))
		if err != nil {
			return errors.New(fmt.Sprintf("Unable to decode properties when trying to register a class %s: %v", name, resp))
		}
		propNames = append(propNames, prop)
	}
	(*place)[name] = regClassEntry{propNames, "@RID, " + strings.Join(propNames, ", ")}
	return nil
}

func (c *Connection) RegisterEClass(name string) error {
	return (*c).registerClass(name, &(*c).regEClasses)
}

func (c *Connection) RegisterVClass(name string) error {
	return (*c).registerClass(name, &(*c).regVClasses)
}

/*func (c *Connection) SelectVertexes(class, cond, queryParams string, limit int) ([](*Vertex), error) {
	regEntry, ok := (*c).regVClasses[class]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Attempt to select vertexes of class %v, which is not registered", class))
	}
	comText := fmt.Sprintf("SELECT %s FROM %s%s%s LIMIT %v", regEntry.propList, class, " "+cond, " "+queryParams, limit)
	res, err := (*c).Command(comText)
	var ret [](*Vertex)
	for _, item := range res {
		var v Vertex
		v.Name = forceToString(item["name"])
		if v.Name == "" {
			return nil, errors.New(fmt.Sprintf("Cannot parse %v as a valid Vertex name", item["name"]))
		}
		v.Rid = forceToString(item["RID"])
		if v.Rid == "" {
			return nil, errors.New(fmt.Sprintf("Cannot parse %v as a valid Vertex RID", item["RID"]))
		}
		v.Data = make(map[string]string)
		for label, prop := range item {
			if label[:1] == "@" || label == "name" || label == "RID" {
				continue
			}
			v.Data[label] = forceToString(prop)
		}
		v.Class = class
		ret = append(ret, &v)
		fmt.Printf("%+v\n", v)
	}
	return nil, err
}*/
