# sheikh

Sheikh is a quick-and-dirty [OrientDB](http://orientdb.com/) driver for Golang. I wrote it for a project of mine,
which didn't make it past the prototype stage. You're free to do with it what you want (it's MIT-licensed). You can
let me know if you'll run into some issues.

## Requirement

Sheikh requires [chillson](https://github.com/szmer/chillson) package for JSON processing.

## Testing

`go test` should work on fresh OrientDB installation (it uses the example database).

## Docs

### Type Connection
```go
type Connection struct {
    Server, Database, Port string
    Username, Password     string
    Client                 http.Client
    // contains filtered or unexported fields
}
```

```go
func NewConnection(servAddr, dbName, user, pass string) (c Connection)
```

NewConnection returns Connection object, which should be initialized with Connect() method before
being used. You have to change the port manually if you wish to:

    c.Port = "8080"

For example,

    c := NewConnection("localhost", "GratefulDeadConcerts", "admin", "admin")

creates a connection to example database shipped with OrientDB
installation.

```go
func (c *Connection) Batch(text string) string
```

```go
func (c *Connection) Command(text string) ([]interface{}, error)
```

Command is a low-level method that performs OrientDB SQL command given in the argument. It returns ["result"] array from JSON
response from the server, which should contain records returned by the
database convertable, to map[string]interface{}. First database error
encountered is copied to the error message of the method.

```go
func (c *Connection) Connect() error
```
Connect method tries to connect to the OrientDB server and perform
authorization.

```go
func (c *Connection) DeleteEdges(rids ...string) error
```
DeleteEdge removes Edge(s) of requested RID(s) from the database.

```go
func (c *Connection) DeleteVertexes(rids ...string) error
DeleteEdge removes Vertex(es) of requested RID(s) from the database.

```go
func (c *Connection) InsertEdge(e *Edge) error
```
InsertEdge inserts given edge to the database, and assings proper RID
and Version values to it.

```go
func (c *Connection) InsertVertex(v *Vertex) error
```
InsertVertex inserts given vertex to the database, and assings proper
RID and Version values to it.

```go
func (c *Connection) SelectEdges(target string, limit int, queryParams string) ([](*Edge), error)
```

SelectEdges returns a slice of Edges from the database. Target is usually a class, but also can be RID. Pass zero or
negative limit if you don't wish to specify maximum number of rows.
queryParams are added verbatim to the underlying SELECT query; it
contain e.g. a WHERE condition.

```go
func (c *Connection) SelectVertexes(target string, limit int, queryParams string) ([](*Vertex), error)
```

SelectVertexes returns a slice of Vertexes from the database. Target is usually a class, but also can be RID. Pass zero or
negative limit if you don't wish to specify maximum number of rows.
queryParams are added verbatim to the underlying SELECT query; it
contain e.g. a WHERE condition.

```go
func (c *Connection) UpdateEdge(e *Edge) error
```

UpdateEdge updates properties of an edge which were changed with SetProp() function since the last sync with
database. Note it silently returns when no changes to the edge were
made. List of changes won't be cleared if any error will be encountered.

```go
func (c *Connection) UpdateVertex(v *Vertex) error
```

UpdateVertex updates properties of a vertex which were changed with SetProp() function since the last sync with
database. Note it silently returns when no changes to the vertex were
made. List of changes won't be cleared if any error will be encountered.

### Type Doc
```go
type Doc struct {
    Class, Rid string // RID should not be specified for the local objects, not uploaded to the db
    Version    int    // version of the object stored in the database
    // contains filtered or unexported fields
}
```
Type Doc contains common object logic of Vertexes and Edges.

### Type Edge
```go
type Edge struct {
    Entry Doc
    // contains filtered or unexported fields
}
```
Type Edge represents OrientDB vertexes (descendants of builtin E class).

```go
func CreateEdge(from *Vertex, className string, to *Vertex) (e Edge)
```

CreateEdge returns an Edge object representing relation between two vertexes of given class; edge must
be inserted to the database before it will be accesible from vertexes'
Edges method.

```go
func (e *Edge) From(c *Connection) (*Vertex, error)
```

From returns Vertex when the Edge starts ("out" Vertex).

```go
func (e Edge) Prop(name string) (interface{}, error)
```
Prop extracts edge's property as {}interface (provided that it is
defined for the Edge).

```go
func (e Edge) PropArr(name string) ([]interface{}, error)
```
PropArr extracts edge's property an array (Go type []interface{})
(provided that it is defined for the Edge).

```go
func (e Edge) PropBool(name string) (bool, error)
```
PropBool extracts edge's property as boolean (provided that it is
defined for the Edge).

```go
func (e Edge) PropFloat(name string) (float64, error)
```
PropFloat extracts edge's property aa float64 (provided that it is
defined for the Edge).

```go
func (e Edge) PropInt(name string) (int, error)
```
PropInt extracts edge's property as int (provided that it is defined for
the Edge).

```go
func (e Edge) PropObj(name string) (map[string]interface{}, error)
```
PropObj extracts edge's property as an object (Go type
map[string]interface{}) (provided that it is defined for the Edge).

```go
func (e Edge) PropRequireArr(name string) []interface{}

func (e Edge) PropRequireBool(name string) bool

func (e Edge) PropRequireFloat(name string) float64

func (e Edge) PropRequireInt(name string) int

func (e Edge) PropRequireObj(name string) map[string]interface{}

func (e Edge) PropRequireStr(name string) string
```

```go
func (e Edge) PropStr(name string) (string, error)
```
PropStr extracts edge's property as string (provided that it is defined
for the Edge).

```go
func (e Edge) RequireProp(name string) interface{}
```

```go
func (e *Edge) SetProps(a ...interface{}) error
```

SetProp takes an arbitrary number of property labels followed by their values. E.g.
SetProp("foo", "bar", "baz", 5) assigns "bar" to "foo" property and 5 to
"baz" property. Method performs assignment in given order, and
terminates if property label is not a string. Arguments are not checked
against schema constraints, which is left to the database.

```go
func (e *Edge) To(c *Connection) (*Vertex, error)
```

From returns Vertex when the Edge ends ("in" Vertex).

### Type EdgeDirection
```go
type EdgeDirection byte
```
    EdgeDirection can be In our Out; Both matches both.

```go
const (
    In EdgeDirection = iota
    Out
    Both
    None
)
```

```go
func (ed EdgeDirection) String() string
```

### Type Vertex
```go
type Vertex struct {
    Entry Doc
    // contains filtered or unexported fields
}
```
Type Vertex represents OrientDB vertexes (descendants of builtin V
class).

```go
func NewVertex(className string) (v Vertex)
```

NewVertex performs essential initialization for Vertex variables, which will not behave
correctly when not created with this function. Vertex can be then
uploaded to the database with Connection.InsertVertex method.

```go
func (v *Vertex) Edges(dirn EdgeDirection,
    with *Vertex,
    className string,
    c *Connection) (ret [](*Edge), err error)
```

Edges returns edges/has that given Vertex has.

```go
func (v Vertex) Prop(name string) (interface{}, error)
```
Prop extracts vertex' property as {}interface (provided that it is
defined for the Vertex).

```go
func (v Vertex) PropArr(name string) ([]interface{}, error)
```
PropArr extracts vertex' property an array (Go type []interface{})
(provided that it is defined for the Vertex).

```go
func (v Vertex) PropBool(namv string) (bool, error)
```
PropBool extracts edge's property as boolean (provided that it is
defined for thv Vertex).

```go
func (v Vertex) PropFloat(name string) (float64, error)
```
PropFloat extracts vertex' property a float64 (provided that it is
defined for thv Vertex).

```go
func (v Vertex) PropInt(name string) (int, error)
```
PropInt extracts vertex' property as int (provided that it is defined
for the Vertex).

```go
func (v Vertex) PropObj(name string) (map[string]interface{}, error)
```
PropObj extracts vertex' property as an object (Go type
map[string]interface{}) (provided that it is defined for the Vertex).

```go
func (v Vertex) PropRequire(name string) interface{}

func (v Vertex) PropRequireArr(name string) []interface{}

func (v Vertex) PropRequireBool(namv string) bool

func (v Vertex) PropRequireFloat(name string) float64

func (v Vertex) PropRequireInt(name string) int

func (v Vertex) PropRequireObj(name string) map[string]interface{}

func (v Vertex) PropRequireStr(name string) string
```

```go
func (v Vertex) PropStr(name string) (string, error)
```
PropStr extracts vertex' property as string (provided that it is defined
for the Vertex).

```go
func (v *Vertex) SetProps(a ...interface{}) error
```

SetProp takes a arbitrary number of property labels followed by their values. E.g.
SetProp("foo", "bar", "baz", 5) assigns "bar" to "foo" property and 5 to
"baz" property. Method performs assignment in given order, and
terminates if property label is not a string. Arguments are not checked
against schema constraints, which is left to the database.


