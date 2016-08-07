package sheikh

import (
	"fmt"
	"os"
	"testing"
)

var c Connection

func TestMain(m *testing.M) {
	c = NewConnection("localhost", "GratefulDeadConcerts", "admin", "admin")
	err := c.Connect()
	if err != nil {
		fmt.Printf("Cannot connect to the database:\n%v\n", err)
		os.Exit(1)
	}
	for i := 0; i < 1; i++ {
		cleanFuncs := [](func() error){
			func() error {
				_, err := c.Command("DROP CLASS Gopher UNSAFE")
				return err
			},
			func() error {
				_, err := c.Command("DROP CLASS owes UNSAFE")
				return err
			},
		}
		prepFuncs := [](func() error){
			func() error {
				_, err := c.Command("CREATE CLASS Gopher EXTENDS V")
				return err
			},
			func() error {
				_, err := c.Command("CREATE PROPERTY Gopher.name string")
				return err
			},
			func() error {
				_, err := c.Command("CREATE CLASS owes EXTENDS E")
				return err
			},
			func() error {
				_, err := c.Command("CREATE PROPERTY owes.howmuch integer")
				return err
			},
		}
		for _, fn := range cleanFuncs { // run them regardless, as there may be some mess left by previous run
			fn()
		}
		for _, fn := range prepFuncs {
			if err := fn(); err != nil {
				fmt.Printf("Test preparation failed:\n%v\n", err)
				os.Exit(1)
			}
		}
		for _, fn := range cleanFuncs {
			if err := fn(); err != nil {
				fmt.Printf("Test cleanup failed:\n%v\n", err)
				os.Exit(1)
			}
		}
	}
}

func TestVertexBasics(t *testing.T) {
	v1 := NewVertex("Gopher")
	err := v1.SetProps("name", "Suzie")
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	err = v1.SetProps("name", "Sue")
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	str, err := v1.PropStr("name")
	if str != "Sue" || err != nil {
		t.Errorf(fmt.Sprintf("GetPropStr returns %v, should be \"Sue\" (error: %v)", str, err))
		return
	}

	v2 := NewVertex("Gopher")
	err = v2.SetProps("name", "Bob")
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	err = v2.SetProps("name", "Mike", "not serious")
	if err == nil {
		t.Errorf("SetProps accepts uneven number of arguments.")
		return
	}
	str, err = v2.PropStr("name")
	if str != "Bob" || err != nil {
		t.Errorf(fmt.Sprintf("GetPropStr returns %v, should be \"Bob\" (error: %v)", str, err))
		return
	}

	err = c.InsertVertex(&v1)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	err = c.InsertVertex(&v2)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
}

func TestEdgeBasics(t *testing.T) {
	vs, err := c.SelectVertexes("Gopher", 2, "")
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	if len(vs) != 2 {
		t.Errorf(fmt.Sprintf("SelectVertexes: Received %v Gopher vertex instances from db, should be 2", len(vs)))
		return
	}
	e := CreateEdge(vs[0], "owes", vs[1])
	err = e.SetProps("howmuch", 111.0)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	money, err := e.PropInt("howmuch")
	if err != nil || money != 111 {
		t.Errorf(fmt.Sprintf("GetPropInt returns %v, should be 111 (error: %v)", money, err))
		return
	}
	err = c.InsertEdge(&e)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
}

func TestUpdates(t *testing.T) {
	vs, err := c.SelectVertexes("Gopher", -1, "WHERE name = \"Sue\"")
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	if len(vs) != 1 {
		t.Errorf(fmt.Sprintf("SelectVertexes: Received %v Gopher vertex instances from db, should be 1", len(vs)))
		return
	}
	vs[0].SetProps("name", "Mary")
	err = c.UpdateVertex(vs[0])
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	if vs[0].Entry.Version != 3 {
		t.Errorf(fmt.Sprintf("Version of the modified vertex appears to be %v, should be 3", vs[0].Entry.Version))
		return
	}
	vs, err = c.SelectVertexes("Gopher", -1, "WHERE name = \"Mary\"")
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	if len(vs) != 1 {
		t.Errorf(fmt.Sprintf("SelectVertexes: Received %v modified Gopher vertex instances from db, should be 1", len(vs)))
		return
	}
}

func TestRelations(t *testing.T) {
	es, err := c.SelectEdges("owes", -1, "")
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	if len(es) != 1 {
		t.Errorf(fmt.Sprintf("SelectEdges: Received %v owes edge instances from db, should be 1", len(es)))
		return
	}
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	v, err := es[0].From(&c)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	v, err = es[0].To(&c)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	es2, err := v.Edges(In, nil, "owes", &c)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	if len(es2) != 1 {
		t.Errorf(fmt.Sprintf("Vertex.Edges: received %v owes edge instances from db, should be 1", len(es2)))
		return
	}
	if es2[0].Entry.Rid == "" || es2[0].Entry.Rid != es[0].Entry.Rid {
		t.Errorf(fmt.Sprintf("RID of selected edge %v and edge retrieved from vertex %v doesn't match", es[0].Entry.Rid, es2[0].Entry.Rid))
		return
	}
}
