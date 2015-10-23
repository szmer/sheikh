package main

import (
	"chillson"
	"errors"
	"fmt"
	"strings"
)

/*type Entity interface {
	Prop(string) (interface{}, error)
	PropArr(string) ([]interface{}, error)
	PropFloat(string) (float64, error)
	PropInt(string) (int, error)
	PropObj(string) (map[string]interface{}, error)
	PropStr(string) (string, error)
	SetProps(...interface{}) error
}*/

type edgeDirection byte

const (
	In edgeDirection = iota
	Out
	Both
)

func (ed edgeDirection) String() string {
	if ed == In {
		return "in"
	}
	if ed == Out {
		return "out"
	}
	return "both"
}

type doc struct {
	Class, Rid     string
	Version        int
	diff           []string // changes since the last update to/from database
	propsContainer map[string]interface{}
	props          chillson.Son
}

type vtxRel struct {
	edgeRid string
}

type Vertex struct {
	Entry doc
	// Maps from edge class names to slices of RIDs.
	edges map[edgeDirection](map[string]([]vtxRel))
}

type Edge struct {
	Entry  doc
	vertex map[edgeDirection]string
}

func docInit(d *doc) {
	d.propsContainer = make(map[string]interface{})
	d.props = chillson.Son{d.propsContainer}
}

/* newEdge returns new Edge. From and to arguments should be RIDs of vertexes forming
the edge. Edges by end user should be rather by creating relation between Vertex objects. */
func newEdge() (e Edge) {
	docInit(&e.Entry)
	e.vertex = make(map[edgeDirection]string)
	return
}

func NewVertex() (v Vertex) {
	v.edges = make(map[edgeDirection](map[string]([]vtxRel)))
	v.edges[In], v.edges[Out] = make(map[string][]vtxRel), make(map[string][]vtxRel)
	docInit(&v.Entry)
	return v
}

func (e Edge) Prop(name string) (interface{}, error) {
	return e.Entry.props.Get("[" + name + "]")
}

func (e Edge) PropArr(name string) ([]interface{}, error) {
	return e.Entry.props.GetArr("[" + name + "]")
}

func (e Edge) PropFloat(name string) (float64, error) {
	return e.Entry.props.GetFloat("[" + name + "]")
}

func (e Edge) PropInt(name string) (int, error) {
	return e.Entry.props.GetInt("[" + name + "]")
}

func (e Edge) PropObj(name string) (map[string]interface{}, error) {
	return e.Entry.props.GetObj("[" + name + "]")
}

// PropStr extracts edge's property as string (provided that it is defined for the Edge).
func (e Edge) PropStr(name string) (string, error) {
	return e.Entry.props.GetStr("[" + name + "]")
}

func (v Vertex) Prop(name string) (interface{}, error) {
	return v.Entry.props.Get("[" + name + "]")
}

func (v Vertex) PropArr(name string) ([]interface{}, error) {
	return v.Entry.props.GetArr("[" + name + "]")
}

func (v Vertex) PropFloat(name string) (float64, error) {
	return v.Entry.props.GetFloat("[" + name + "]")
}

func (v Vertex) PropInt(name string) (int, error) {
	return v.Entry.props.GetInt("[" + name + "]")
}

func (v Vertex) PropObj(name string) (map[string]interface{}, error) {
	return v.Entry.props.GetObj("[" + name + "]")
}

// PropStr extracts vertex' property as string (provided that it is defined for the Vertex).
func (v Vertex) PropStr(name string) (string, error) {
	return v.Entry.props.GetStr("[" + name + "]")
}

func setProps(container *map[string]interface{}, diff *[]string, a []interface{}) error {
	if len(a) == 0 && len(a)%2 != 0 {
		return errors.New("SetProp: no arguments or odd number of arguments")
	}
	for i := 0; i < len(a); i += 2 {
		label, ok := a[i].(string)
		if !ok {
			return errors.New(fmt.Sprintf("SetProp: non-string label %v", a[i]))
		}
		val, _ := a[i+1].(interface{})
		(*container)[label] = val
		(*diff) = append(*diff, label)
	}
	return nil
}

/* SetProp takes a arbitrary number of property labels followed by their values. E.g.
SetProp("foo", "bar",  "baz", 5) assigns "bar" to "foo" property and 5 to "baz" property.
Method performs assignment in given order, and terminates if property label is not a string.
Arguments are not checked against schema constraints, which is left to the database. */
func (e *Edge) SetProps(a ...interface{}) error {
	return setProps(&e.Entry.propsContainer, &e.Entry.diff, a)
}

/* SetProp takes a arbitrary number of property labels followed by their values. E.g.
SetProp("foo", "bar",  "baz", 5) assigns "bar" to "foo" property and 5 to "baz" property.
Method performs assignment in given order, and terminates if property label is not a string.
Arguments are not checked against schema constraints, which is left to the database. */
func (v *Vertex) SetProps(a ...interface{}) error {
	return setProps(&v.Entry.propsContainer, &v.Entry.diff, a)
}

func CreateEdge(from *Vertex, className string, to *Vertex) (e Edge) {
	e = newEdge()
	e.Entry.Class = className
	e.vertex[Out] = from.Entry.Rid
	e.vertex[In] = to.Entry.Rid
	return
}

func (e *Edge) From(c *Connection) (*Vertex, error) {
	if (*c).vertexes[e.vertex[Out]] != nil {
		return (*c).vertexes[e.vertex[Out]], nil
	}
	vs, err := (*c).SelectVertexes(e.vertex[Out], 1, "", "")
	if err != nil {
		return nil, err
	}
	if len(vs) != 1 {
		return nil, errors.New(fmt.Sprintf("Edge out vertex of RID %s cannot be found in database", e.vertex[Out]))
	}
	return vs[0], nil
}

func (e *Edge) To(c *Connection) (*Vertex, error) {
	if (*c).vertexes[e.vertex[In]] != nil {
		return (*c).vertexes[e.vertex[In]], nil
	}
	vs, err := (*c).SelectVertexes(e.vertex[In], 1, "", "")
	if err != nil {
		return nil, err
	}
	if len(vs) != 1 {
		return nil, errors.New(fmt.Sprintf("Edge in vertex of RID %s cannot be found in database", e.vertex[In]))
	}
	return vs[0], nil
}

func (v *Vertex) Edges(dirn edgeDirection,
	with *Vertex,
	className string,
	c *Connection) (ret [](*Edge), err error) {
	// At least for now, we handle cases with "both" directions with recurency.
	if dirn == Both {
		in, err := v.Edges(In, with, className, c)
		if err != nil {
			return in, err
		}
		out, err := v.Edges(Out, with, className, c)
		ret = append(in, out...)
		return ret, err
	}
	var aggregate relSliceAggregate
	if className != "" {
		aggregate = NewRelSliceAggregate(v.edges[dirn][className], nil)
	} else {
		aggregate = NewRelSliceAggregate(nil, v.edges[dirn])
	}
	missingRids := make([]string, 0)
	if with == nil { // relation to any other vertex
		for relEntry := aggregate.yield(); relEntry.edgeRid != ""; relEntry = aggregate.yield() {
			edge := c.edges[relEntry.edgeRid]
			if edge == nil {
				missingRids = append(missingRids, relEntry.edgeRid)
				continue
			}
			ret = append(ret, edge)
		}
	} else { // relation to some prescribed vertex
		for relEntry := aggregate.yield(); relEntry.edgeRid != ""; relEntry = aggregate.yield() {
			edge := c.edges[relEntry.edgeRid]
			if edge == nil {
				missingRids = append(missingRids, relEntry.edgeRid)
				continue
			}
			if edge.vertex[dirn] == v.Entry.Rid {
				ret = append(ret, edge)
			}
		}
	}
	if len(missingRids) == 0 {
		return ret, nil
	}
	queryTarget := "[" + strings.Join(missingRids, ",") + "]"
	var queryCond string
	if with != nil {
		queryCond = fmt.Sprintf("WHERE %s = %s", dirn, with.Entry.Rid)
	}
	missingEdges, err := c.SelectEdges(queryTarget, 0, queryCond, "")
	if err != nil {
		return ret, err
	}
	ret = append(ret, missingEdges...)
	return ret, nil
}
