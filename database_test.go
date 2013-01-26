package tohva

import (
  "testing"
  "log"
)

// in these tests we assume again that there exists a user named `admin` with password `admin`
// who is database administrator

func TestCreateDatabase(t *testing.T) {
  couch := CreateCouchClient(TestCouchHost, 5984)
  session := couch.StartSession()
  session.Login("admin", "admin")

  db := session.GetDatabase("tohva_test")

  if db.Exists() {
    db.Delete()
  }

  if !db.Create() {
    t.Error("pffffff! you didn't clean the couchdb instance fuckin' bastard")
  } else {
    // cleaning
    log.Println(db.Delete())
  }
}

func TestGetInfo(t *testing.T) {
  couch := CreateCouchClient(TestCouchHost, 5984)
  session := couch.StartSession()
  session.Login("admin", "admin")

  db := session.GetDatabase("tohva_test")

  if(db.Exists()) {
    if db.GetInfo() == nil {
      t.Error("I just tried, and it existed. Is somebody else playing with your couchdb instance? Then, kill him and try again")
    }
    db.Delete()
    if db.GetInfo() != nil {
      t.Error("You gotta be kidding me... I just deleted the database for you")
    }
  } else {
    if db.GetInfo() != nil {
      t.Error("I just tried, and it did not exist. Is somebody else playing with your couchdb instance? Then, kill him and try again")
    }
    db.Create()
    if db.GetInfo() == nil {
      t.Error("You gotta be kidding me... I just created the database for you")
    }
  }
}
