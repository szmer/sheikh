package gorient

import (
	"chillson"
	"errors"
	"fmt"
	"strings"
)

/* DeleteEdge removes Edge(s) of requested RID(s) from the database. */
func (c *Connection) DeleteEdges(rids ...string) error {
	comText := fmt.Sprintf("DELETE EDGE %s", strings.Join(rids, ","))
	_, err := (*c).Command(comText)
	return err
}

/* DeleteEdge removes Vertex(es) of requested RID(s) from the database. */
func (c *Connection) DeleteVertexes(rids ...string) error {
	comText := fmt.Sprintf("DELETE VERTEX %s", strings.Join(rids, ","))
	_, err := (*c).Command(comText)
	return err
}

func (c *Connection) insertEntry(entry *Doc, entryComText string) error {
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
	comText := fmt.Sprintf("CREATE EDGE %s FROM %s TO %s", (*e).Entry.Class, e.vertex[Out], e.vertex[In])
	err := c.insertEntry(&e.Entry, comText)
	if err != nil {
		return err
	}
	if vtx, present := c.vertexes[e.vertex[Out]]; present {
		vtx.edges[Out][e.Entry.Class] = append(vtx.edges[Out][e.Entry.Class], vtxRel{e.Entry.Rid})
	}
	if vtx, present := c.vertexes[e.vertex[In]]; present {
		vtx.edges[In][e.Entry.Class] = append(vtx.edges[In][e.Entry.Class], vtxRel{e.Entry.Rid})
	}
	return nil
}

/* InsertVertex inserts given vertex to the database, and assings proper RID and Version values to it.*/
func (c *Connection) InsertVertex(v *Vertex) error {
	comText := fmt.Sprintf("CREATE VERTEX %s", (*v).Entry.Class)
	return c.insertEntry(&v.Entry, comText)
}

func unpackProps(entry *Doc, origEntry interface{}) (err error) {
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

/* SelectEdges returns a slice of Edges from the database. Target is usually a class, but also can be RID. Pass zero or
negative limit if you don't wish to specify maximum number of rows. queryParams are added verbatim to the underlying SELECT
query; it contain e.g. a WHERE condition. */
func (c *Connection) SelectEdges(target string, limit int, queryParams string) ([](*Edge), error) {
	comText := fmt.Sprintf("SELECT FROM %s%s", target, queryParams)
	if limit > 1 {
		comText += fmt.Sprintf(" LIMIT %v", limit)
	}
	res, err := (*c).Command(comText)
	var ret [](*Edge)
	for ind := range res {
		e := newEdge()
		err = unpackProps(&e.Entry, res[ind]) // TODO: break on err?
		e.vertex[Out], err = e.PropStr("out")
		if err == nil {
			e.vertex[In], err = e.PropStr("in")
		}
		if err != nil { // serious business
			return nil, errors.New(fmt.Sprintf("SelectEdges: edge cannot be read properly, error: %v", err))
		}
		delete(e.Entry.propsContainer, "out")
		delete(e.Entry.propsContainer, "in")
		c.edges[e.Entry.Rid] = &e // add to the index
		ret = append(ret, &e)
	}
	return ret, err
}

/* SelectVertexes returns a slice of Vertexes from the database. Target is usually a class, but also can be RID. Pass zero or
negative limit if you don't wish to specify maximum number of rows. queryParams are added verbatim to the underlying SELECT
query; it contain e.g. a WHERE condition. */
func (c *Connection) SelectVertexes(target string, limit int, queryParams string) ([](*Vertex), error) {
	comText := fmt.Sprintf("SELECT FROM %s%s", target, " "+queryParams)
	if limit > 1 {
		comText += fmt.Sprintf(" LIMIT %v", limit)
	}
	res, err := (*c).Command(comText)
	var ret [](*Vertex)
	for ind := range res {
		v := NewVertex("")
		err = unpackProps(&v.Entry, res[ind]) // TODO: break on err?
		var (                                 // for processing edges/relations when they're encountered
			relClass string
			relDirn  EdgeDirection
		)
		for label, val := range v.Entry.propsContainer {
			if label[:4] == "out_" && len(label) > 4 {
				relClass, relDirn = label[4:], Out
				goto ParseRelations
			}
			if label[:3] == "in_" && len(label) > 3 {
				relClass, relDirn = label[3:], In
				goto ParseRelations
			}
			continue

		ParseRelations:
			rels, ok := val.([]interface{})
			if !ok {
				return ret, errors.New(fmt.Sprintf("SelectVertexes: Cannot process edges of type %s", relClass))
			}
			v.edges[relDirn][relClass] = nil // initialize
			for _, rawEdgeRid := range rels {
				edgeRid := rawEdgeRid.(string)
				if !ok {
					return ret, errors.New(fmt.Sprintf("SelectVertexes: Cannot process edges of type %s", relClass))
				}
				v.edges[relDirn][relClass] = append(v.edges[relDirn][relClass], vtxRel{edgeRid})
			}
			if err != nil { // error when parsing edges/relations
				return ret, err
			}
			delete(v.Entry.propsContainer, label)
		}
		c.vertexes[v.Entry.Rid] = &v // add to the index
		ret = append(ret, &v)
	}
	return ret, err
}

func (c *Connection) updateEntry(entry *Doc) error {
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
