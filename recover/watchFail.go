package recover

import (
	"encoding/json"
	"time"

	"github.com/Sirupsen/logrus"
	"jvole.com/influx/db"
	"jvole.com/influx/route"
	"jvole.com/influx/util"
)

type WatchFail interface {
	//根据本地数据写回influxdb
	WriteRec()
	//开启进程监控本地数据和influxdb
	Run()
	//监控influx状态
	WatchServer()
}

const (
	NUM      = 10000 //每次回复10000条
	Interval = 10    //检查时间间隔
)

type watchFail struct {
	server []route.Server
}

//返回实例
func NewWatchFail() watchFail {
	wf := &watchFail{}
	wf.SetServer()
	// for _, v := range wf.server { //循环服务器
	// 	pp.Println(v.Addr)
	// 	// pp.Println(str)
	// }

	return *wf
}

func (wf *watchFail) SetServer() {
	wf.server = nil
	for _, v := range route.Groups { //组
		for _, sv := range v.S { //每组的服务器
			wf.server = append(wf.server, sv)
		}
	}
	// for _, v := range wf.server { //循环服务器
	// 	pp.Println(v.Addr)
	// 	// pp.Println(str)
	// }
}

func (wf *watchFail) Run() {
	go func() {
		for range time.Tick(time.Duration(Interval) * time.Second) {
			wf.WatchServer()
			wf.WriteRec()
		}
	}()
}

//根据本地数据恢复到influxdb
func (wf *watchFail) WriteRec() {
	var data struct {
		Addr   string
		Tags   map[string]string
		Fields map[string]interface{}
		Table  string
		Time   time.Time
	}
	var sdata struct { //服务器本地数据
		Key   string
		Value string
	}

	for _, v := range wf.server { //循环服务器
		if !v.Status { //服务器状态不可用后面不执行
			continue
		}
		indb := db.NewInfluxdb(v.Addr, v.Username, v.Password, v.Dbname, v.Precision)
		str := db.RecordE.Search(v.Addr, NUM)
		// str := db.RecordE.Search("https", NUM)

		for _, sv := range str { //循环此服务器的失败数据

			json.Unmarshal([]byte(sv), &sdata)
			json.Unmarshal([]byte(sdata.Value), &data)
			// pp.Println(data)
			err := indb.Insert(data.Tags, data.Fields, data.Table, data.Time)
			if err == nil { //插入成功删除本地数据
				db.RecordE.Delete(sdata.Key)
			} else {
				_, _, err = indb.Ping(time.Second * 1)
				if err == nil { //检查服务器是否正常，正常说明数据格式等错误，删除数据
					db.RecordE.Delete(sdata.Key)
				}
				util.Log.WithFields(logrus.Fields{
					"name": "回复数据",
					"err":  err,
				}).Infoln("回复数据错误")
			}
		}
		indb.Close()
	}
}

func (wf *watchFail) WatchServer() {
	for i := 0; i < len(wf.server); i++ { //循环服务器
		indb := db.NewInfluxdb(wf.server[i].Addr, wf.server[i].Username, wf.server[i].Password, wf.server[i].Dbname, wf.server[i].Precision)
		_, _, err := indb.Ping(time.Second * 1) //检查服务器状态
		if err != nil {
			// util.Log.WithFields(logrus.Fields{
			// 	"name":   "错误",
			// 	"err":    err,
			// 	"server": wf.server[i].Addr,
			// }).Debugln("服务器连接失败")
			wf.server[i].Status = false //改变此服务器状态
		} else {
			wf.server[i].Status = true //改变此服务器状态
		}
		indb.Close()
	}
}
