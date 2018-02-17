package bidder

import (
	"errors"
	_ "fmt"

	"github.com/influxdata/influxdb/client/v2"
	indb "jvole.com/influx/db"
)

type serverBidder struct {
	influxDB indb.Influxdb
}

//错误定义
var (
	ErrEmpty = errors.New("empty string")
)

func (s *serverBidder) Query(cmd, db, precision string, limit, offset int) (res []client.Result, err error) {
	return s.influxDB.QueryDB(cmd, db, precision, limit, offset)
}
func (s *serverBidder) Insert(tags map[string]string, fields map[string]interface{}, table string) error {
	return s.influxDB.InsertWithBuffer(tags, fields, table)
}
func (s *serverBidder) InsertBatch(data []indb.IndbData, table string) error {
	return s.influxDB.InsertBatch(data, table)
}
func (s *serverBidder) InsertNow(tags map[string]string, fields map[string]interface{}, table string) error {
	err := s.influxDB.InsertWithBuffer(tags, fields, table)
	if err != nil {
		return err
	}
	err = s.influxDB.InsertNow()
	return err
}

/*
返回服务实例
@param addr influx地址
@param user 用户名
@param password 密码
@param db 数据库名
@param precision 精确度
@param buffer 缓存数量
@return serverBidder struct
*/
func NewServer(addr, user, password, db, precision string, buffer uint16) *serverBidder {
	sb := &serverBidder{}
	sb.influxDB = indb.NewInfluxdb(addr, user, password, db, precision)

	indb.Buffer = buffer
	return sb
}
