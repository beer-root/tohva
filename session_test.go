package tohva

import "testing"

func TestGetSessionInfo(t *testing.T) {
  couch := CreateCouchClient(TestCouchHost, 5984)
  session := couch.StartSession()
  t.Log(session.GetSessionInfo())
}

func TestLogin(t *testing.T) {
  couch := CreateCouchClient(TestCouchHost, 5984)
  session := couch.StartSession()
  if !session.Login("admin", "admin") {
    t.Error("wrong admin credentials")
  }
  if session.Login("admin", "truie") {
    t.Error("these credentials should not be correct")
  }
}

func TestIsLoggedIn(t *testing.T) {
  couch := CreateCouchClient(TestCouchHost, 5984)
  session := couch.StartSession()
  if session.IsLoggedIn() {
    t.Error("w00t?? how is that even possible")
  }
  if !session.Login("admin", "admin") {
    t.Error("wrong admin credentials")
  }
  if !session.IsLoggedIn() {
    t.Error("yes but you should be logged in you moron")
  }
}

func TestLogout(t *testing.T) {
  couch := CreateCouchClient(TestCouchHost, 5984)
  session := couch.StartSession()
  if session.IsLoggedIn() {
    t.Error("w00t?? how is that even possible")
  }
  if !session.Login("admin", "admin") {
    t.Error("wrong admin credentials")
  }
  if !session.IsLoggedIn() {
    t.Error("yes but you should be logged in you moron")
  }
  session.Logout()
  if session.IsLoggedIn() {
    t.Error("yes but you should be logged out now")
  }
}
