package tohva

import (
  "strconv"
  "net/http"
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
}

type CouchUsers struct {
  DbName string
}

type CouchSession struct {
  CouchDB
  Cookie string
}

// creates a couchdb client
func CreateCouchClient(host string, port int) CouchDB {
  return CouchDB{ host, port, "1.2", CouchUsers{ "_users" } }
}

// start a new session for this couchdb instance
func (couch CouchDB) StartSession() CouchSession {
  return CouchSession { couch, "" }
}

func (couch CouchDB) url() string {
  return couch.Host + ":" + (strconv.Itoa(couch.Port))
}

// get the base request for this couchdb instance
func (couch CouchDB) NewRequest(method string) (*http.Request, error) {
  url := couch.url()
  return http.NewRequest(method, url, nil)
}

// ========== Couch Session ==========

func (session CouchSession) Login(name string, password string) bool {
  return true
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
  Id string `json: "_id"`
  Rev string `json: "_rev,omitempty"`
  Language string `json:"language"`
  Views map[string] View `json:"views,omitempty"`
  ValidateDocUpdate string `json: "validate_doc_update,omitempty"`
}

type View struct {
  Map string `json:"map"`
  Reduce string `json:"map,omitempty"`
}

// ========== Internals ==========

type simpleResult struct {
  Ok bool `json: "ok"`
  Id string `json: "id,omitempty"`
  Rev string `json: "rev,omitempty"`
}

func getIdAndRev(doc interface{}) 
