package main

import (
	"chillson"
	"errors"
	"fmt"
)

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

func (c *Connection) DeleteVertex(rid string) error {
	comText := fmt.Sprintf("DELETE VERTEX %s", rid)
	_, err := (*c).Command(comText)
	return err
}

func (c *Connection) SelectVertexes(class string, limit int, cond, queryParams string) ([](*Vertex), error) {
	comText := fmt.Sprintf("SELECT FROM %s%s%s LIMIT %v", class, " "+cond, " "+queryParams, limit)
	res, err := (*c).Command(comText)
	var ret [](*Vertex)
	for ind := range res {
		chill := chillson.Son{res[ind]}
		var v Vertex
		v.Class = class
		v.Rid, err = chill.GetStr("[@rid]")
		if err == nil {
			v.Version, err = chill.GetInt("[@version]")
		}
		if err != nil {
			return nil, err
		}
		props, _ := res[ind].(map[string]interface{})
		delete(props, "@rid") // delete duplicate properties
		delete(props, "@version")
		delete(props, "@class")
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
