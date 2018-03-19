package route

import (
	"errors"
	_ "fmt"
	"time"

	"github.com/Sirupsen/logrus"
	indb "jvole.com/influx/db"
	"jvole.com/influx/util"
)

type serverRoute struct {
	dbGateway InfluxGateway
}

//错误定义
var (
	ErrEmpty = errors.New("empty string")
)

func (s serverRoute) Insert(tags map[string]string, fields map[string]interface{}, table string, uid uint64) error {
	var err error
	num, gp := s.dbGateway.GetGroup(uid)
	for _, v := range gp.S {
		util.Log.WithFields(logrus.Fields{
			"uid": uid,
			"服务器": v.Conn.ToString(),
			"组":   num,
		}).Debugln("调试信息")
		err = v.Conn.InsertWithBuffer(tags, fields, table)
		if err != nil {
			util.Log.WithFields(logrus.Fields{
				"err":    err,
				"server": v.Addr,
				"group":  num,
				"table":  table,
				"tags":   tags,
				"fields": fields,
			}).Infoln("写数据库失败")
		}
	}
	return err
}
func (s serverRoute) InsertNow(tags map[string]string, fields map[string]interface{}, table string, uid uint64) error {
	var err error
	num, gp := s.dbGateway.GetGroup(uid)
	for _, v := range gp.S {
		util.Log.WithFields(logrus.Fields{
			"uid": uid,
			"服务器": v.Conn.ToString(),
			"组":   num,
		}).Debugln("调试信息")
		err = v.Conn.InsertWithBuffer(tags, fields, table)
		if err != nil {
			util.Log.WithFields(logrus.Fields{
				"err":    err,
				"server": v.Addr,
				"group":  num,
				"table":  table,
				"tags":   tags,
				"fields": fields,
			}).Infoln("写数据库失败")
			return err
		}
		err = v.Conn.InsertNow()
		if err != nil {
			util.Log.WithFields(logrus.Fields{
				"err":    err,
				"server": v.Addr,
				"group":  num,
				"table":  table,
				"tags":   tags,
				"fields": fields,
			}).Infoln("写数据库失败")
		}

	}
	return err
}
func (s serverRoute) InsertBatch(data []indb.IndbData, table string, uid uint64) error {
	var err error
	num, gp := s.dbGateway.GetGroup(uid)
	for _, v := range gp.S {
		util.Log.WithFields(logrus.Fields{
			"uid": uid,
			"服务器": v.Conn.ToString(),
			"组":   num,
		}).Debugln("调试信息")
		err = v.Conn.InsertBatch(data, table)
		if err != nil {
			util.Log.WithFields(logrus.Fields{
				"err":    err,
				"server": v.Addr,
				"group":  num,
				"table":  table,
				"data":   data,
			}).Infoln("写数据库失败")
		}
	}
	return err
}

func (s serverRoute) Delete(cmd, db, precision string, uid uint64) (err error) {
	num, gp := s.dbGateway.GetGroup(uid)
	for _, v := range gp.S {
		util.Log.WithFields(logrus.Fields{
			"uid": uid,
			"服务器": v.Conn.ToString(),
			"组":   num,
		}).Debugln("调试信息")
		err = v.Conn.Delete(cmd, db, precision)
		if err != nil {
			util.Log.WithFields(logrus.Fields{
				"err":    err,
				"server": v.Addr,
				"group":  num,
				"cmd":    cmd,
				"db":     db,
			}).Infoln("写数据库失败")
			return
		}
	}
	return
}

func (s serverRoute) Select(cmd, db, precision string, limit, offset int, uid []uint64) (res []QueryResult, err error) {
	for i := 0; i < len(uid); i++ { //遍历所有id
		num, gp := s.dbGateway.GetGroup(uid[i])
		var conn indb.Influxdb
		for i := 0; i < len(gp.S)*5; i++ { //检查服务器状态
			conn = gp.Poll.Next().(indb.Influxdb)
			_, _, err = conn.Ping(time.Second * 2)
			// pp.Println(conn.ToString())
			if err == nil {
				break
			}
		}
		util.Log.WithFields(logrus.Fields{
			"uid": uid[i],
			"服务器": conn.ToString(),
			"组":   num,
		}).Debugln("调试信息")
		//查询
		result, err := conn.Select(cmd, db, precision, limit, offset)

		if err != nil {
			util.Log.WithFields(logrus.Fields{
				"err":   err,
				"group": num,
			}).Infoln("查询失败")
		} else {
			r := QueryResult{uid[i], result}
			res = append(res, r)
		}
	}
	return
}

func (s serverRoute) Query(cmd, db, precision string, uid []uint64) (res []QueryResult, err error) {
	for i := 0; i < len(uid); i++ { //遍历所有id
		num, gp := s.dbGateway.GetGroup(uid[i])
		var conn indb.Influxdb
		for i := 0; i < len(gp.S)*5; i++ { //检查服务器状态
			conn = gp.Poll.Next().(indb.Influxdb)
			_, _, err = conn.Ping(time.Second * 2)
			// pp.Println(conn.ToString())
			if err == nil {
				break
			}
		}
		util.Log.WithFields(logrus.Fields{
			"uid": uid[i],
			"服务器": conn.ToString(),
			"组":   num,
		}).Debugln("调试信息")
		//查询
		result, err := conn.Query(cmd, db, precision)

		if err != nil {
			util.Log.WithFields(logrus.Fields{
				"err":   err,
				"group": num,
			}).Error("执行命令失败")
			return res, err
		} else {
			r := QueryResult{uid[i], result}
			res = append(res, r)
		}
	}

	return
}

/*
返回服务实例
x
@return serverRoute struct
*/
func NewServer() *serverRoute {
	sb := &serverRoute{}
	sb.dbGateway = &Group{}
	return sb
}
