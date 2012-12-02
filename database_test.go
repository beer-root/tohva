package tohva

import "testing"

// in these tests we assume again that there exists a user named `admin` with password `admin`
// who is database administrator

func TestCreateDatabase(t * testing.T) {
  couch := CreateCouchClient("localhost", 5984)
  session := couch.StartSession()
  session.Login("admin", "admin")

  db := session.GetDatabase("tohva_test")

  if !db.Create() {
    t.Error("pffffff! you didn't clean the couchdb instance fuckin' bastard")
  } else {
    // cleaning
    db.Delete()
  }
}
