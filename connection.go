package main

import (
	"chillson"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

type Connection struct {
	Server, Database, Port string
	Username, Password     string
	client                 http.Client
	// Index of vertexes and edges received from the db (indexed by RIDs).
	vertexes map[string](*Vertex)
	edges    map[string](*Edge)
}

func NewConnection(serv, db, user, pass string) (c Connection) {
	c.Server = serv
	c.Database = db
	c.Username = user
	c.Password = pass

	c.vertexes = make(map[string](*Vertex))
	c.edges = make(map[string](*Edge))

	c.Port = "2480"

	c.client.Jar, _ = cookiejar.New(nil)
	return
}

func (c *Connection) Command(text string) ([]interface{}, error) {
	text = url.QueryEscape(text)
	addr := fmt.Sprintf("http://%s:%s/command/%s/sql/%s", (*c).Server, (*c).Port, (*c).Database, text)
	req, err := http.NewRequest("POST", addr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept-Encoding", "gzip,deflate")
	req.Header.Set("Content-Length", "0")

	resp, err := (*c).client.Do(req)
	buff := make([]byte, 10240)
	p, err := io.ReadFull(resp.Body, buff)
	if err != io.ErrUnexpectedEOF {
		resp.Body.Close()
		return nil, err
	}
	//TODO: how to act when buffer exceeded
	resp.Body.Close()
	buff = buff[:p]

	var respJson interface{}
	json.Unmarshal(buff, &respJson)
	chill := chillson.Son{respJson}
	firstErr, err := chill.GetObj("[errors][0]")
	if err == nil {
		chill = chillson.Son{firstErr} // extract from ['errors'][0]
		reason, _ := chill.GetStr("[reason]")
		content, _ := chill.GetStr("[content]")
		return nil, errors.New(fmt.Sprintf("Command %v failed, server error reason: %v; content: %q", text, reason, content))
	}
	result, err := chill.GetArr("[result]")
	if err == nil {
		return result, nil
	}
	return nil, errors.New(fmt.Sprintf("Unable to extract result from server response to command %v, response body: %v", text, respJson))
}

func (c *Connection) Connect() error {
	addr := fmt.Sprintf("http://%s:%s/connect/%s", (*c).Server, (*c).Port, (*c).Database)
	req, err := http.NewRequest("GET", addr, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth((*c).Username, (*c).Password)

	resp, err := (*c).client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 204 {
		return errors.New(fmt.Sprintf("Connecting to OrientDB: HTTP status %v, perhaps wrong credentials", resp.Status))
	}
	if cookies := (*c).client.Jar.Cookies(req.URL); len(cookies) != 1 || strings.Index(cookies[0].String(), "OSESSIONID=") == -1 {
		return errors.New("Connecting to OrientDB: connection ok, but OSESSIONID cookie not present in server response, wrong address?")
	}
	return err // nil if all OK
}

func main() {
	c := NewConnection("localhost", "GratefulDeadConcerts", "admin", "admin")
	err := c.Connect()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	vs, err := c.SelectVertexes("#17:3", 0, "", "")
	if err != nil {
		fmt.Printf("bład: %v\n", err)
		return
	}
	es, err := vs[0].Edges(Out, nil, "lubi", &c)
	if err != nil {
		fmt.Printf("błąd: %v\n", err)
		return
	}
	for _, v := range es {
		fmt.Printf("%+v\n", *v)
	}
}
