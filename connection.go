package sheikh

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
	"time"
)

type Connection struct {
	Server, Database, Port string
	Username, Password     string
	Client                 http.Client
	// Index of vertexes and edges received from the db (indexed by RIDs).
	vertexes map[string](*Vertex)
	edges    map[string](*Edge)
}

/* NewConnection returns Connection object, which should be initialized with Connect() method before
being utilized. You have to change the port manually if you wish to:
   c.Port = "8080"

For example,
   c := NewConnection("localhost", "GratefulDeadConcerts", "admin", "admin")
creates a connection to example database shipped with OrientDB installation.*/
func NewConnection(servAddr, dbName, user, pass string) (c Connection) {
	c.Server = servAddr
	c.Database = dbName
	c.Username = user
	c.Password = pass

	c.vertexes = make(map[string](*Vertex))
	c.edges = make(map[string](*Edge))

	c.Port = "2480"

	c.Client.Jar, _ = cookiejar.New(nil)
	c.Client.Timeout = 2 * time.Second
	return
}

type respAndError struct {
	resp *http.Response
	err  error
}

// doRequest spawns a goroutine, which should do a request, and handles timeout.
func (c *Connection) doRequest(req *http.Request) (*http.Response, error) {
	requestDone := make(chan respAndError)
	go func() {
		resp, err := (*c).Client.Do(req)
		requestDone <- respAndError{resp, err}
		return
	}()
	var result respAndError
	result = <-requestDone
	return result.resp, result.err
}

/* Command is a low-level method that performs OrientDB SQL command given in the argument. It returns ["result"] array from JSON
response from the server, which should contain records returned by the database convertable, to map[string]interface{}. First database
error encountered is copied to the error message of the method. */
func (c *Connection) Command(text string) ([]interface{}, error) {
	text = url.QueryEscape(text)
	addr := fmt.Sprintf("http://%s:%s/command/%s/sql/%s", (*c).Server, (*c).Port, (*c).Database, text)
	req, err := http.NewRequest("POST", addr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept-Encoding", "gzip,deflate")
	req.Header.Set("Content-Length", "0")

	resp, err := (*c).doRequest(req)
	if err != nil {
		return nil, err
	}
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

/* Connect method tries to connect to the OrientDB server and perform authorization. */
func (c *Connection) Connect() error {
	addr := fmt.Sprintf("http://%s:%s/connect/%s", (*c).Server, (*c).Port, (*c).Database)
	req, err := http.NewRequest("GET", addr, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth((*c).Username, (*c).Password)

	resp, err := (*c).doRequest(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 204 {
		return errors.New(fmt.Sprintf("Connecting to OrientDB: HTTP status %v, perhaps wrong credentials", resp.Status))
	}
	if cookies := (*c).Client.Jar.Cookies(req.URL); len(cookies) != 1 || strings.Index(cookies[0].String(), "OSESSIONID=") == -1 {
		return errors.New("Connecting to OrientDB: connection ok, but OSESSIONID cookie not present in server response, wrong address?")
	}
	return err // nil if all OK
}
