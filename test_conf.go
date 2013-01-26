package tohva

import "code.google.com/p/goconf/conf"

var TestCouchHost string = initHost()

func initHost() string {
  c, err := conf.ReadConfigFile("./tohva-test.conf")
  if err == nil {
    h, err := c.GetString("database", "host")
    if err == nil {
      return h
    }
  }
  return "localhost"
}
