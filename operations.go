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
	entry.propList = "@RID, @version"
	for i, _ := range sliceProps {
		var prop struct {
			name    string
			odbType float64
		}
		prop.name, err = chill.GetStr(fmt.Sprintf("[%v][name]", i))
		if err == nil {
			prop.odbType, err = chill.GetFloat(fmt.Sprintf("[%v][type]", i))
		}
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

/* InsertVertex inserts given vertex to the database, and assings proper RID and Version values to it.*/
func (c *Connection) InsertVertex(v *Vertex) error {
	comText := fmt.Sprintf("INSERT INTO %s", (*v).Class)
	specialProps := false
	for label, prop := range (*v).propsContainer {
		if !specialProps {
			comText += " SET "
			specialProps = true
		}
		comText += fmt.Sprintf(" %s = %v", label, toOdbRepr(prop))
	}
	ret, err := (*c).Command(comText)
	fmt.Printf("%v\n", ret)
	chill := chillson.Son{ret}
	(*v).Rid, err = chill.GetStr("[0][@rid]")
	(*v).Version = 1
	return err
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
		v.Rid, err = chill.GetStr("[RID]")
		if err == nil {
			v.Version, err = chill.GetInt("[version]")
		}
		if err != nil {
			return nil, err
		}
		props, _ := res[ind].(map[string]interface{})
		delete(props, "RID")
		delete(props, "version")
		for key := range props {
			if key[:1] == "@" {
				delete(props, key)
			}
		}
		v.propsContainer = props
		v.props = chillson.Son{v.propsContainer}
		ret = append(ret, &v)
	}
	return ret, err
}

/* UpdateVertex updates properties of a vector which were changed with SetProp() function since the last sync with
database. Note it silently returns when no changes to the vector were made. List of changes won't be cleared if any
error will be encountered. */
func (c *Connection) UpdateVertex(v *Vertex) error {
	if v.diff == nil {
		return nil
	}
	if v.Rid == "" {
		return errors.New("UpdateVertex: it has no associated RID, did vertex come from the db?")
	}
	comText := fmt.Sprintf("UPDATE %s SET", v.Rid)
	for _, label := range v.diff {
		comText += fmt.Sprintf(" %s = %s", label, toOdbRepr(v.propsContainer[label]))
	}
	comText += " RETURN AFTER @version"
	resp, err := (*c).Command(comText)
	if err != nil {
		return err
	}
	chill := chillson.Son{resp}
	v.Version, _ = chill.GetInt("[0][value]")
	v.diff = nil
	return nil
}
