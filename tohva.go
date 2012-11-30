package tohva

import (
  "fmt"
  "io"
  "io/ioutil"
  "encoding/json"
  "strconv"
  "strings"
  "net/http"
  "net/url"
  "log"
  "time"
)

// this file contains the interface to the couchdb server
// this is more or less a port of the sohva project for Scala
// https://github.com/gnieh/sohva

// ========== Couch Instance ==========

type CouchDB struct {
  Host string
  Port int
  Version string
  Users CouchUsers
  // internals
  client http.Client
}

type CouchUsers struct {
  DbName string
}

type CouchSession struct {
  *CouchDB
  Cookie string
}

// creates a couchdb client
func CreateCouchClient(host string, port int) CouchDB {
  return CouchDB{ host, port, "1.2", CouchUsers{ "_users" }, http.Client{} }
}

// send a request that accepts a json and parse its result to the expected type
func (couch *CouchDB) doJsonRequest(method string, path string, body io.Reader, result interface{}) error {
  req, err := couch.newRequest(method, path, body)
  if err != nil {
    return err
  }
  err = req.ParseForm();
  if err != nil {
    return err
  }
  // set the accept header
  req.Header.Add("Accept", "application/json")
  // send the request
  resp, err := couch.client.Do(req)
  if err != nil {
    return err
  }
  // parse the result
  respData, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return err
  }
  return json.Unmarshal(respData, result)
}

// the cookie jar
type cookieJar struct {
  couchCookie *http.Cookie
}

// TODO make it thread safe
func (jar *cookieJar) SetCookies(url *url.URL, cookies []*http.Cookie) {
  for i := range cookies {
    if cookies[i].Name == "AuthSession" {
      jar.couchCookie = cookies[i]
    }
  }
}

// TODO make it thread safe
func (jar *cookieJar) Cookies(url *url.URL) []*http.Cookie {
  return []*http.Cookie{jar.couchCookie}
}

// start a new session for this couchdb instance
func (couch *CouchDB) StartSession() CouchSession {
  // add my personal cookie jar for this session
  // copy the underlying couch client
  my_couch := *couch
  my_couch.client.Jar = &cookieJar{nil}
  return CouchSession { &my_couch, "" }
}

func (couch CouchDB) url() string {
  return couch.Host + ":" + (strconv.Itoa(couch.Port) + "/")
}

// get the base request for this couchdb instance
func (couch CouchDB) newRequest(method string, path string, body io.Reader) (*http.Request, error) {
  url := couch.url() + path
  return http.NewRequest(method, url, body)
}

// ========== Couch Session ==========

func (session *CouchSession) Login(name string, password string) bool {
  var r loginResult
  body := strings.NewReader(fmt.Sprintf("name=%s&password=%s", name, password))
  if err := session.doJsonRequest("POST", "_session", body, &r); err != nil {
    log.Println("[ERROR]", time.Now(), err)
    return false
  }

  return r.Ok
}

func (session *CouchSession) Logout() bool {
  var r simpleResult
  if err := session.doJsonRequest("DELETE", "_session", nil, &r); err != nil {
    log.Println("[ERROR]", time.Now(), err)
    return false
  }

  return r.Ok
}

// ========== Database ==========

type Database struct {
  Name string
  couch CouchDB
}

// returns an object that allows user to work with a couchdb database
func (couch CouchDB) GetDatabase(name string) Database {
  return Database { name, couch }
}

func (d Database) SaveDesign(design Design) (string, string, error) {
  return "", "", nil
}

// ========== Designs and Views ==========

type Design struct {
  Id string `json:"_id"`
  Rev string `json:"_rev,omitempty"`
  Language string `json:"language"`
  Views map[string] View `json:"views,omitempty"`
  ValidateDocUpdate string `json:"validate_doc_update,omitempty"`
}

type View struct {
  Map string `json:"map"`
  Reduce string `json:"map,omitempty"`
}

// ========== Internals ==========

type simpleResult struct {
  Ok bool `json:"ok"`
  Id string `json:"id,omitempty"`
  Rev string `json:"rev,omitempty"`
}

type errorResult struct {
  Error string `json:"error"`
  Reason string `json:"reason"`
}

type loginResult struct {
  Ok bool `json:"ok"`
  Name string `json:"name"`
  Roles []string `json:"roles"`
}

func getIdAndRev(doc interface{})
