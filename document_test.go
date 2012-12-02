package tohva

import "testing"

type Doc struct {
  WithIdRev
  Value string `json:"value"`
}

func TestSaveIn(t *testing.T) {
  couch := CreateCouchClient("localhost", 5984)
  session := couch.StartSession()
  session.Login("admin", "admin")

  db := session.GetDatabase("tohva_test")

  if !db.Exists() {
    db.Create()
  }

  doc := &Doc{WithIdRev{"test_doc", nil}, "toto"}

  if db.SaveDoc(doc) != nil {
    t.Error("I should be able to save the document...")
  }

  if doc.Rev == nil {
    t.Error("The revision should not be null")
  }

  db.Delete()

}
