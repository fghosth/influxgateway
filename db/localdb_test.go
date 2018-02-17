package db_test

import (
	"testing"

	"github.com/k0kubun/pp"
	"jvole.com/influx/db"
)

var conn = db.NewBoltDB()

// func TestSave(t *testing.T) {
// 	conn.Save("key", "value")
// 	fmt.Println(conn.Load("key"))
// }
//
// func TestLoad(t *testing.T) {
// 	conn.Save("key2", "value4")
// 	fmt.Println(conn.Load("key2"))
// }

func TestSearch(t *testing.T) {
	// for i := 0; i < 100; i++ {
	// 	key := "http://localhost:8989||" + strconv.Itoa(i)
	// 	value := `{"tags":{"server":"ddd"},"fields":{"name":"jack"}}`
	// 	conn.Save(key, value)
	// }
	str := conn.Search("https", 10)

	// str := db.Load("https://localhost:8086||a1412f7706e21f76ed2a27cd912f52a9")
	pp.Println(str)
}

// func TestDelete(t *testing.T) {
// 	// for i := 0; i < 10000; i++ {
// 	// 	key := "derek||" + strconv.Itoa(i)
// 	// 	value := `{"tags":{"server":"ddd"},"fields":{"name":"jack"}}`
// 	// 	conn.Save(key, value)
// 	// }
// 	var sdata struct {
// 		Key   string
// 		Value string
// 	}
// 	str := conn.Search("derek", 20)
// 	for _, v := range str {
// 		json.Unmarshal([]byte(v), &sdata)
// 		// conn.Delete(sdata.Key)
// 		pp.Println(sdata)
// 	}
//
// }
