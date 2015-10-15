package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
)

type Connection struct {
	Server, Database, Port string
	Username, Password     string
	Ok                     bool
	client                 http.Client

	regEClasses map[string]regClassEntry
	regVClasses map[string]regClassEntry
}

type regClassEntry struct {
	props    []string
	propList string
}

type Json map[string]interface{}

func NewConnection(serv, db, user, pass string) (c Connection) {
	c.Server = serv
	c.Database = db
	c.Username = user
	c.Password = pass

	c.Port = "2480"
	c.Ok = false

	c.client.Jar, _ = cookiejar.New(nil)
	c.regEClasses = make(map[string]regClassEntry)
	c.regVClasses = make(map[string]regClassEntry)
	return
}

func (c *Connection) Command(text string) ([]Json, error) {
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
		return nil, err
	}
	//TODO: how to act when buffer exceeded
	resp.Body.Close()
	buff = buff[:p]

	respJson := Json(make(map[string]interface{}))
	var respJsonPnt *Json = &respJson

	err = json.Unmarshal(buff, respJsonPnt)
	if respErr, present := respJson["errors"]; present {
		respErrSlice, ok := respErr.([]interface{})
		var respFirstErr Json // declared here to let program do goto jump
		if !ok || len(respErrSlice) < 1 {
			goto CommandCannotDecodeError
		}
		respFirstErr, ok = respErrSlice[0].(map[string]interface{})
		if !ok {
			goto CommandCannotDecodeError
		}
		return nil, errors.New(fmt.Sprintf("Command %v failed, server error reason: %v; content: %q", text, respFirstErr["reason"], respFirstErr["content"]))
	CommandCannotDecodeError:
		return nil, errors.New(fmt.Sprintf("Command %v failed, unable to decode server error JSON: %v", text, respErr))
	}
	if respResult, present := respJson["result"]; present {
		rawSlice, ok := respResult.([]interface{})
		var jsonSlice []Json
		if !ok {
			goto CommandCannotDecodeResult
		}
		for _, rawElem := range rawSlice {
			jsonElem, ok := rawElem.(map[string]interface{})
			if !ok {
				goto CommandCannotDecodeResult
			}
			jsonSlice = append(jsonSlice, jsonElem)
		}
		return jsonSlice, nil
	}
CommandCannotDecodeResult:
	return nil, errors.New(fmt.Sprintf("Unable to extract result from server response to command %v, response body: %v", text, respJson))
}

func (c *Connection) Connect() error {
	//TODO: blockout when server is down?
	addr := fmt.Sprintf("http://%s:%s/connect/%s", (*c).Server, (*c).Port, (*c).Database)
	req, err := http.NewRequest("GET", addr, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth((*c).Username, (*c).Password)

	resp, err := (*c).client.Do(req)
	defer resp.Body.Close()
	if resp.StatusCode != 204 {
		return errors.New(fmt.Sprintf("Connecting to OrientDB: HTTP status %v, perhaps wrong credentials", resp.Status))
	}
	if cookies := (*c).client.Jar.Cookies(req.URL); len(cookies) != 1 || strings.Index(cookies[0].String(), "OSESSIONID=") == -1 {
		return errors.New("Connecting to OrientDB: connection ok, but OSESSIONID cookie not present in server response, wrong address?")
	}
	return err // nil if all OK
}

func (c *Connection) registerClass(name string, place *map[string]regClassEntry) error {
	var propNames []string
	var sliceProps []interface{}
	resp, err := (*c).Command(fmt.Sprintf("SELECT classes[name='%s'] FROM metadata:schema", name))
	if err != nil {
		return err
	}
	classesMap, ok := resp[0]["classes"].(map[string]interface{})
	if !ok {
		goto RegisterCannotDecodeProp
	}
	sliceProps, ok = classesMap["properties"].([]interface{})
	if !ok {
		goto RegisterCannotDecodeProp
	}
	for _, rawProp := range sliceProps {
		prop, ok := rawProp.(map[string]interface{})
		if !ok {
			goto RegisterCannotDecodeProp
		}
		propName, ok := prop["name"].(string)
		if !ok {
			goto RegisterCannotDecodeProp
		}
		propNames = append(propNames, propName)
	}
	(*place)[name] = regClassEntry{propNames, "@RID, " + strings.Join(propNames, ", ")}
	return nil
RegisterCannotDecodeProp:
	return errors.New(fmt.Sprintf("Unable to decode properties when trying to register a class %s: %v", name, resp))

}

func (c *Connection) RegisterEClass(name string) error {
	return (*c).registerClass(name, &(*c).regEClasses)
}

func (c *Connection) RegisterVClass(name string) error {
	return (*c).registerClass(name, &(*c).regVClasses)
}

func (c *Connection) SelectVertexes(class, cond, queryParams string, limit int) ([](*Vertex), error) {
	regEntry, ok := (*c).regVClasses[class]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Attempt to select vertexes of class %v, which is not registered", class))
	}
	comText := fmt.Sprintf("SELECT %s FROM %s%s%s LIMIT %v", regEntry.propList, class, " "+cond, " "+queryParams, limit)
	res, err := (*c).Command(comText)
	var ret [](*Vertex)
	for _, item := range res {
		var v Vertex
		v.Name = forceToString(item["name"])
		if v.Name == "" {
			return nil, errors.New(fmt.Sprintf("Cannot parse %v as a valid Vertex name", item["name"]))
		}
		v.Rid = forceToString(item["RID"])
		if v.Rid == "" {
			return nil, errors.New(fmt.Sprintf("Cannot parse %v as a valid Vertex RID", item["RID"]))
		}
		v.Data = make(map[string]string)
		for label, prop := range item {
			if label[:1] == "@" || label == "name" || label == "RID" {
				continue
			}
			v.Data[label] = forceToString(prop)
		}
		v.Class = class
		ret = append(ret, &v)
		fmt.Printf("%+v\n", v)
	}
	return nil, err
}

func main() {
	c := NewConnection("localhost", "GratefulDeadConcerts", "admin", "admin")
	err := c.Connect()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	err = c.RegisterVClass("ORole")
	_, err = c.SelectVertexes("ORole", "", "", 10)
	/*resp, err := c.Command("select classes[name='OUser'] from metadata:schema")
	fmt.Printf("%v\n", resp)*/
}
