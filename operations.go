package main

import (
	"chillson"
	"errors"
	"fmt"
	"strings"
)

func (c *Connection) DeleteEdge(rid string) error {
	comText := fmt.Sprintf("DELETE EDGE %s", rid)
	_, err := (*c).Command(comText)
	return err
}

func (c *Connection) DeleteVertex(rid string) error {
	comText := fmt.Sprintf("DELETE VERTEX %s", rid)
	_, err := (*c).Command(comText)
	return err
}

func (c *Connection) insertEntry(entry *doc, entryComText string) error {
	specialProps := false
	for label, prop := range (*entry).propsContainer {
		if !specialProps {
			entryComText += " SET "
			specialProps = true
		}
		entryComText += fmt.Sprintf(" %s = %v", label, toOdbRepr(prop))
	}
	ret, err := (*c).Command(entryComText)
	chill := chillson.Son{ret}
	(*entry).Rid, err = chill.GetStr("[0][@rid]")
	(*entry).Version = 1
	return err
}

/* InsertEdge inserts given edge to the database, and assings proper RID and Version values to it.*/
func (c *Connection) InsertEdge(e *Edge) error {
	comText := fmt.Sprintf("CREATE EDGE %s FROM %s TO %s", (*e).Entry.Class, e.fromRid, e.toRid)
	return c.insertEntry(&e.Entry, comText)
}

/* InsertVertex inserts given vertex to the database, and assings proper RID and Version values to it.*/
func (c *Connection) InsertVertex(v *Vertex) error {
	comText := fmt.Sprintf("CREATE VERTEX %s", (*v).Entry.Class)
	return c.insertEntry(&v.Entry, comText)
}

func unpackProps(entry *doc, origEntry interface{}) (err error) {
	chill := chillson.Son{origEntry}
	(*entry).Class, err = chill.GetStr("[@class]")
	if err == nil {
		(*entry).Rid, err = chill.GetStr("[@rid]")
	}
	if err == nil {
		(*entry).Version, err = chill.GetInt("[@version]")
	}
	if err != nil {
		return err
	}
	props, _ := origEntry.(map[string]interface{})
	delete(props, "@rid") // delete duplicate properties
	delete(props, "@version")
	delete(props, "@class")
	(*entry).propsContainer = props
	(*entry).props = chillson.Son{(*entry).propsContainer}
	return err
}

// target is usually a class, but also can be RID
func (c *Connection) SelectEdges(target string, limit int, cond, queryParams string) ([](*Edge), error) {
	comText := fmt.Sprintf("SELECT FROM %s%s%s LIMIT %v", target, " "+cond, " "+queryParams, limit)
	res, err := (*c).Command(comText)
	var ret [](*Edge)
	for ind := range res {
		var e Edge
		err = unpackProps(&e.Entry, res[ind]) // TODO: break on err?
		e.fromRid, err = e.PropStr("out")
		if err == nil {
			e.toRid, err = e.PropStr("in")
		}
		if err != nil { // serious business
			return nil, errors.New(fmt.Sprintf("SelectEdges: edge cannot be read properly, error: %v", err))
		}
		delete(e.Entry.propsContainer, "out")
		delete(e.Entry.propsContainer, "in")
		ret = append(ret, &e)
	}
	return ret, err
}

func (c *Connection) SelectVertexes(target string, limit int, cond, queryParams string) ([](*Vertex), error) {
	comText := fmt.Sprintf("SELECT FROM %s%s%s LIMIT %v", target, " "+cond, " "+queryParams, limit)
	res, err := (*c).Command(comText)
	var ret [](*Vertex)
	for ind := range res {
		v := Vertex{}.init()
		err = unpackProps(&v.Entry, res[ind]) // TODO: break on err?
		var (                                 // variables for processing edges/relations when they're encountered
			relClass string
			relIndex *map[string][]vtxRel
		)
		for label, val := range v.Entry.propsContainer {
			if label[:4] == "out_" && len(label) > 4 {
				relClass, relIndex = label[4:], &v.out
				goto ParseRelations
			}
			if label[:3] == "in_" && len(label) > 3 {
				relClass, relIndex = label[3:], &v.in
				goto ParseRelations
			}
			continue

		ParseRelations:
			rels, ok := val.([]interface{})
			if !ok {
				return ret, errors.New(fmt.Sprintf("SelectVertexes: Cannot process edges of type %s", relClass))
			}
			(*relIndex)[relClass] = nil // initialize
			for _, rawEdgeRid := range rels {
				edgeRid := rawEdgeRid.(string)
				if !ok {
					return ret, errors.New(fmt.Sprintf("SelectVertexes: Cannot process edges of type %s", relClass))
				}
				(*relIndex)[relClass] = append((*relIndex)[relClass], vtxRel{edgeRid, nil})
			}
			if err != nil { // error when parsing edges/relations
				return ret, err
			}
			delete(v.Entry.propsContainer, label)
		}
		ret = append(ret, &v)
	}
	return ret, err
}

func (c *Connection) updateEntry(entry *doc) error {
	if (*entry).Rid == "" {
		return errors.New("Update: entity has no associated RID, did it come from the db?")
	}
	if (*entry).diff == nil {
		return nil
	}
	comText := fmt.Sprintf("UPDATE %s SET", (*entry).Rid)
	var removeList []string
	for _, label := range (*entry).diff {
		val, present := (*entry).propsContainer[label]
		if !present {
			removeList = append(removeList, label)
			continue
		}
		comText += fmt.Sprintf(" %s = %s", label, toOdbRepr(val))
	}
	if len(removeList) != 0 {
		comText += " REMOVE " + strings.Join(removeList, " ")
	}
	comText += " RETURN AFTER @version"
	resp, err := (*c).Command(comText)
	if err != nil {
		return err
	}
	chill := chillson.Son{resp}
	(*entry).Version, err = chill.GetInt("[0][value]")
	(*entry).diff = nil
	return err
}

/* UpdateEdge updates properties of an edge which were changed with SetProp() function since the last sync with
database. Note it silently returns when no changes to the edge were made. List of changes won't be cleared if any
error will be encountered. */
func (c *Connection) UpdateEdge(e *Edge) error {
	return c.updateEntry(&e.Entry)
}

/* UpdateVertex updates properties of a vertex which were changed with SetProp() function since the last sync with
database. Note it silently returns when no changes to the vertex were made. List of changes won't be cleared if any
error will be encountered. */
func (c *Connection) UpdateVertex(v *Vertex) error {
	return c.updateEntry(&v.Entry)
}
