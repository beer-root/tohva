package tohva

import (
  "io"
  "io/ioutil"
  "fmt"
  "encoding/json"
  "strconv"
  "net/http"
  "net/url"
  "strings"
  "log"
  "sync"
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
func (couch *CouchDB) doJsonRequest(method string, path string, body io.Reader, form bool, result interface{}) error {
  req, err := couch.newRequest(method, path, body)
  if err != nil {
    return err
  }
  // set the accept header
  req.Header.Add("Accept", "application/json")
  // url encoded form submission
  if form {
    req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
    req.Header.Add("Referer", "http://localhost:5984")
  }
	if couch.client.Jar != nil {
		for _, cookie := range couch.client.Jar.Cookies(req.URL) {
			req.AddCookie(cookie)
		}
	}
  // send the request
  resp, err := couch.client.Do(req)
  if err != nil {
    return err
  } else if err == nil && couch.client.Jar != nil {
		couch.client.Jar.SetCookies(req.URL, resp.Cookies())
	}
  // don't forget to close the bodies
  if req.Body != nil {
    defer req.Body.Close()
  }
  defer resp.Body.Close()
  respData, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return err
  }
  // parse the result
  if resp.StatusCode / 200 == 1 {
    // status code is 2XX -> ok
    if len(respData) > 0 && result != nil {
      err = json.Unmarshal(respData, result)
      if err != nil {
        return err
      }
    }
    return nil
  }

  var couchErr CouchError
  couchErr.StatusCode = resp.StatusCode

  if len(respData) > 0 {
    err = json.Unmarshal(respData, &couchErr)
    if err != nil {
      return err
    }
  }

  return couchErr
}

// the cookie jar
type cookieJar struct {
  couchCookie *http.Cookie
  lk sync.Mutex
}

var EmptyCookie = http.Cookie{Name: "AuthSession", Value: ""}

func (jar *cookieJar) SetCookies(url *url.URL, cookies []*http.Cookie) {
  jar.lk.Lock()
  defer jar.lk.Unlock()
  for i := range cookies {
    if cookies[i].Name == "AuthSession" {
      jar.couchCookie = cookies[i]
    }
  }
}

func (jar *cookieJar) Cookies(url *url.URL) []*http.Cookie {
  jar.lk.Lock()
  defer jar.lk.Unlock()
  return []*http.Cookie{jar.couchCookie}
}

// start a new session for this couchdb instance
func (couch *CouchDB) StartSession() CouchSession {
  // add my personal cookie jar for this session
  // copy the underlying couch client
  my_couch := *couch
  my_couch.client.Jar = &cookieJar{&EmptyCookie, sync.Mutex{}}
  return CouchSession { &my_couch, "" }
}

func (couch CouchDB) url() string {
  return "http://" + couch.Host + ":" + (strconv.Itoa(couch.Port) + "/")
}

// get the base request for this couchdb instance
func (couch CouchDB) newRequest(method string, path string, body io.Reader) (*http.Request, error) {
  url := couch.url() + path
  return http.NewRequest(method, url, body)
}

// ========== Couch Session ==========

// Log user in (cookie based authentication)
func (session *CouchSession) Login(name string, password string) bool {
  var r loginResult
  v := url.Values{}
  v.Set("name", name)
  v.Set("password", password)
  body := strings.NewReader(v.Encode())
  if err := session.doJsonRequest("POST", "_session", body, true, &r); err != nil {
    log.Println("[ERROR]", err)
    return false
  }

  return r.Ok
}

// Log user out
func (session *CouchSession) Logout() bool {
  var r simpleResult
  if err := session.doJsonRequest("DELETE", "_session", nil, false, &r); err != nil {
    log.Println("[ERROR]", err)
    return false
  }

  return r.Ok
}

type SessionInfo struct {
  Ok bool `json:"ok"`
  Context *SessionContext `json:"userCtx"`
  Info struct {
    Authenticated *string `json:"authenticated,omitempty"`
  } `json:"info"`
}

type SessionContext struct {
  Name *string `json:"name"`
  Roles []string `json:"roles"`
}

// Gets the current session context information
func (session *CouchSession) GetSessionInfo() *SessionInfo {
  var info SessionInfo
  if err := session.doJsonRequest("GET", "_session", nil, false, &info); err != nil {
    log.Println("[ERROR]", err)
    return nil
  }
  return &info
}

// Is user logged in?
func (session *CouchSession) IsLoggedIn() bool {
  info := session.GetSessionInfo()
  return info != nil && info.Info.Authenticated != nil
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

// returns the database url
func (db Database) GetUrl() string {
  return db.couch.url() + db.Name
}

// checks whether the database exists in the couchdb instance
func (db Database) Exists() bool {
  err := db.couch.doJsonRequest("HEAD", db.Name, nil, false, nil)
  if err != nil {
    switch err.(type) {
    case CouchError:
      if err.(CouchError).StatusCode != 404 {
        log.Println("[ERROR]", err)
      }
    default:
      log.Println("[ERROR]", err)
    }
    return false
  }
  return true
}


// creates the database. returns true iff the database was actually created, not if it already existed
func (db Database) Create() bool {
  if !db.Exists() {
    var resp simpleResult
    err := db.couch.doJsonRequest("PUT", db.Name, nil, false, &resp)
    if err != nil {
      log.Println("[ERROR]", err)
      return false
    }
    return resp.Ok
  }
  return false
}

func (db Database) Delete() bool {
  if db.Exists() {
    var resp simpleResult
    err := db.couch.doJsonRequest("DELETE", db.Name, nil, false, &resp)
    if err != nil {
      log.Println("[ERROR]", err)
      return false
    }
    return resp.Ok
  }
  return false
}

// save the document into the database. if it is successfully saved, then the revision is modified in place
func (db Database) SaveDoc(doc *IdRev) *CouchError {
  return &CouchError{}
}

func (db Database) SaveDesign(design *Design) *CouchError {
  return &CouchError{}
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
  Id *string `json:"id,omitempty"`
  Rev *string `json:"rev,omitempty"`
}

type loginResult struct {
  Ok bool `json:"ok"`
  Name *string `json:"name"`
  Roles []string `json:"roles"`
}

// ========== Some Userful Types ==========

type CouchError struct {
  Msg string `json:"error"`
  Reason string `json:"reason"`
  StatusCode int
}

func (err CouchError) Error() string {
  return fmt.Sprintf("%d: %s caused by %s", err.StatusCode, err.Msg, err.Reason)
}

type DbInfo struct {
  Name string `json:"db_name"`
  DocCount int64 `json:"doc_count"`
  DocDelCount int64 `json:"doc_del_count"`
  UpdateSeq int64 `json:"update_seq"`
  PurgeSeq int64 `json:"purge_seq"`
  CompactRunning bool `json:"compact_running"`
  DiskSize int64 `json:"disk_size"`
  DataSize int64 `json:"data_size"`
  InstanceStartTime string `json:"instance_start_time"`
  DiskFormatVersion int64 `json:"disk_format_version"`
  CommittedUpdateSeq int64 `json:"committed_update_seq"`
}

type IdRev struct {
  Id string `json:"_id"`
  Rev *string `json:"_rev,omitempty"`
}
