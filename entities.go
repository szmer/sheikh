package sheikh

import (
	"chillson"
	"errors"
	"fmt"
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

/* EdgeDirection can be In our Out; Both matches both. */
type EdgeDirection byte

const (
	In EdgeDirection = iota
	Out
	Both
	None
)

func (ed EdgeDirection) String() string {
	switch ed {
	case In:
		return "in"
	case Out:
		return "out"
	case Both:
		return "both"
	}
	return "none"
}

/* Type Doc contains common object logic of Vertexes and Edges. */
type Doc struct {
	Class, Rid     string   // RID should not be specified for the local objects, not uploaded to the db
	Version        int      // version of the object stored in the database
	diff           []string // changes since the last update to/from database
	propsContainer map[string]interface{}
	props          chillson.Son
}

type vtxRel struct {
	edgeRid string
}

// Type Vertex represents OrientDB vertexes (descendants of builtin V class).
type Vertex struct {
	Entry Doc
	// Maps from edge class names to slices of RIDs.
	edges map[EdgeDirection](map[string]([]vtxRel))
}

// Type Edge represents OrientDB vertexes (descendants of builtin E class).
type Edge struct {
	Entry  Doc
	vertex map[EdgeDirection]string
}

func docInit(d *Doc) {
	d.propsContainer = make(map[string]interface{})
	d.props = chillson.Son{d.propsContainer}
}

/* CreateEdge returns an Edge object representing relation between two vertexes of given class; edge must
be inserted to the database before it will be accesible from vertexes' Edges method. */
func CreateEdge(from *Vertex, className string, to *Vertex) (e Edge) {
	e = newEdge()
	e.Entry.Class = className
	e.vertex[Out] = from.Entry.Rid
	e.vertex[In] = to.Entry.Rid
	return
}

/* newEdge returns new Edge. From and to arguments should be RIDs of vertexes forming
the edge. Edges by end user should be rather by creating relation between Vertex objects. */
func newEdge() (e Edge) {
	docInit(&e.Entry)
	e.vertex = make(map[EdgeDirection]string)
	return
}

/* NewVertex performs essential initialization for Vertex variables, which will not behave
correctly when not created with this function. Vertex can be then uploaded to the database with
Connection.InsertVertex method. */
func NewVertex(className string) (v Vertex) {
	v.edges = make(map[EdgeDirection](map[string]([]vtxRel)))
	v.edges[In], v.edges[Out] = make(map[string][]vtxRel), make(map[string][]vtxRel)
	docInit(&v.Entry)
	v.Entry.Class = className
	return v
}

// Prop extracts edge's property as {}interface (provided that it is defined for the Edge).
func (e Edge) Prop(name string) (interface{}, error) {
	return e.Entry.props.Get("[" + name + "]")
}

func (e Edge) RequireProp(name string) interface{} {
	return e.Entry.props.Require("[" + name + "]")
}

// PropArr extracts edge's property an array (Go type []interface{}) (provided that it is defined for the Edge).
func (e Edge) PropArr(name string) ([]interface{}, error) {
	return e.Entry.props.GetArr("[" + name + "]")
}

func (e Edge) PropRequireArr(name string) []interface{} {
	return e.Entry.props.RequireArr("[" + name + "]")
}

// PropBool extracts edge's property as boolean (provided that it is defined for the Edge).
func (e Edge) PropBool(name string) (bool, error) {
	return e.Entry.props.GetBool("[" + name + "]")
}

func (e Edge) PropRequireBool(name string) bool {
	return e.Entry.props.RequireBool("[" + name + "]")
}

// PropFloat extracts edge's property aa float64 (provided that it is defined for the Edge).
func (e Edge) PropFloat(name string) (float64, error) {
	return e.Entry.props.GetFloat("[" + name + "]")
}

func (e Edge) PropRequireFloat(name string) float64 {
	return e.Entry.props.RequireFloat("[" + name + "]")
}

// PropInt extracts edge's property as int (provided that it is defined for the Edge).
func (e Edge) PropInt(name string) (int, error) {
	return e.Entry.props.GetInt("[" + name + "]")
}

func (e Edge) PropRequireInt(name string) int {
	return e.Entry.props.RequireInt("[" + name + "]")
}

// PropObj extracts edge's property as an object (Go type map[string]interface{}) (provided that it is defined for the Edge).
func (e Edge) PropObj(name string) (map[string]interface{}, error) {
	return e.Entry.props.GetObj("[" + name + "]")
}

func (e Edge) PropRequireObj(name string) map[string]interface{} {
	return e.Entry.props.RequireObj("[" + name + "]")
}

// PropStr extracts edge's property as string (provided that it is defined for the Edge).
func (e Edge) PropStr(name string) (string, error) {
	return e.Entry.props.GetStr("[" + name + "]")
}

func (e Edge) PropRequireStr(name string) string {
	return e.Entry.props.RequireStr("[" + name + "]")
}

// Prop extracts vertex' property as {}interface (provided that it is defined for the Vertex).
func (v Vertex) Prop(name string) (interface{}, error) {
	return v.Entry.props.Get("[" + name + "]")
}

func (v Vertex) PropRequire(name string) interface{} {
	return v.Entry.props.Require("[" + name + "]")
}

// PropArr extracts vertex' property an array (Go type []interface{}) (provided that it is defined for the Vertex).
func (v Vertex) PropArr(name string) ([]interface{}, error) {
	return v.Entry.props.GetArr("[" + name + "]")
}

func (v Vertex) PropRequireArr(name string) []interface{} {
	return v.Entry.props.RequireArr("[" + name + "]")
}

// PropBool extracts edge's property as boolean (provided that it is defined for thv Vertex).
func (v Vertex) PropBool(namv string) (bool, error) {
	return v.Entry.props.GetBool("[" + namv + "]")
}

func (v Vertex) PropRequireBool(namv string) bool {
	return v.Entry.props.RequireBool("[" + namv + "]")
}

// PropFloat extracts vertex' property a float64 (provided that it is defined for thv Vertex).
func (v Vertex) PropFloat(name string) (float64, error) {
	return v.Entry.props.GetFloat("[" + name + "]")
}

func (v Vertex) PropRequireFloat(name string) float64 {
	return v.Entry.props.RequireFloat("[" + name + "]")
}

// PropInt extracts vertex' property as int (provided that it is defined for the Vertex).
func (v Vertex) PropInt(name string) (int, error) {
	return v.Entry.props.GetInt("[" + name + "]")
}

func (v Vertex) PropRequireInt(name string) int {
	return v.Entry.props.RequireInt("[" + name + "]")
}

// PropObj extracts vertex' property as an object (Go type map[string]interface{}) (provided that it is defined for the Vertex).
func (v Vertex) PropObj(name string) (map[string]interface{}, error) {
	return v.Entry.props.GetObj("[" + name + "]")
}

func (v Vertex) PropRequireObj(name string) map[string]interface{} {
	return v.Entry.props.RequireObj("[" + name + "]")
}

// PropStr extracts vertex' property as string (provided that it is defined for the Vertex).
func (v Vertex) PropStr(name string) (string, error) {
	return v.Entry.props.GetStr("[" + name + "]")
}

func (v Vertex) PropRequireStr(name string) string {
	return v.Entry.props.RequireStr("[" + name + "]")
}

func setProps(container *map[string]interface{}, diff *[]string, a []interface{}) error {
	if len(a) == 0 || len(a)%2 != 0 {
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

/* SetProp takes an arbitrary number of property labels followed by their values. E.g.
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

/* From returns Vertex when the Edge starts ("out" Vertex). */
func (e *Edge) From(c *Connection) (*Vertex, error) {
	if (*c).vertexes[e.vertex[Out]] != nil {
		return (*c).vertexes[e.vertex[Out]], nil
	}
	vs, err := (*c).SelectVertexes(e.vertex[Out], 1, "")
	if err != nil {
		return nil, err
	}
	if len(vs) != 1 {
		return nil, errors.New(fmt.Sprintf("Edge out vertex of RID %s cannot be found in database", e.vertex[Out]))
	}
	return vs[0], nil
}

/* From returns Vertex when the Edge ends ("in" Vertex). */
func (e *Edge) To(c *Connection) (*Vertex, error) {
	if (*c).vertexes[e.vertex[In]] != nil {
		return (*c).vertexes[e.vertex[In]], nil
	}
	vs, err := (*c).SelectVertexes(e.vertex[In], 1, "")
	if err != nil {
		return nil, err
	}
	if len(vs) != 1 {
		return nil, errors.New(fmt.Sprintf("Edge in vertex of RID %s cannot be found in database", e.vertex[In]))
	}
	return vs[0], nil
}

/* Edges returns edges/has that given Vertex has. */
func (v *Vertex) Edges(dirn EdgeDirection,
	with *Vertex,
	className string,
	c *Connection) (ret [](*Edge), err error) {
	queryCond, target := "WHERE", "E"
	if className != "" {
		target = className
	}
	if dirn == In || dirn == Out {
		queryCond += fmt.Sprintf(" %v = %s", dirn, v.Entry.Rid)
	} else {
		queryCond += fmt.Sprintf(" (in = %s OR out = %s)", v.Entry.Rid, v.Entry.Rid)
	}
	if with != nil {
		switch dirn {
		case In:
			queryCond += fmt.Sprintf(" AND out = %s", with.Entry.Rid)
		case Out:
			queryCond += fmt.Sprintf(" AND in = %s", with.Entry.Rid)
		default:
			queryCond += fmt.Sprintf(" AND (in = %s OR out = %s)", with.Entry.Rid)
		}
	}
	return c.SelectEdges(target, -1, queryCond)
}
