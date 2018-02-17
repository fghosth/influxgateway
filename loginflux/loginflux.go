package loginflux

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

//日志插件，可以把日志记录到influxdb，实现了io.writer接口能够用于所有，传入io.writer的log日志插件（支持json格式的）
type Loginflux struct {
	client      client.Client
	batchPoints client.BatchPoints //数据集合
	server      InfluxServer
	mtx         sync.RWMutex
}

type InfluxServer struct {
	Addr      string            //连接地址
	User      string            //用户名
	Passwd    string            //密码
	Database  string            //数据库名
	Table     string            //表名
	Precision string            //精确度
	Buff      uint16            //当前缓存数量
	Tags      map[string]string //存放influxdb时的tags
	KeyS      []string          //如果日志中出现这些字段就把它假如到tags中
}

var (
	PrintScreen = false //是同时否输出到屏幕 true 为显示
)

//实现io.Writer
func (lf *Loginflux) Write(p []byte) (n int, err error) {
	lf.mtx.Lock()
	defer lf.mtx.Unlock()
	var result map[string]interface{}
	if err = json.Unmarshal(p, &result); err != nil {
		return n, err
	}
	for _, v := range lf.server.KeyS { //添加tags
		if val, ok := result[v].(string); ok {
			lf.server.Tags[v] = val
			delete(result, v)
		}
	}
	pt, err := client.NewPoint(lf.server.Table, lf.server.Tags, result, time.Now())
	if err != nil {
		return n, err
	}
	lf.batchPoints.AddPoint(pt)
	if uint16(len(lf.batchPoints.Points())) >= lf.server.Buff { //缓存满了写数据库
		//写数据库
		if err := lf.client.Write(lf.batchPoints); err != nil { //这里会把所有有效数据添加
			fmt.Println("日志写入错误:" + err.Error())
		}
		lf.batchPoints.ClearPoint() //源码增加了这个方法 注意包
	}
	//输出到屏幕
	if PrintScreen {
		os.Stdout.Write(p)
	}

	return n, err

}

func NewLoginflux(is InfluxServer) io.Writer {
	lf := &Loginflux{}
	lf.server = is
	// fmt.Printf("addr:%s,user:%s,pwd:%s\n", addr, user, password)
	// 创建 point batch
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  lf.server.Database,
		Precision: lf.server.Precision,
	})
	if err != nil {
		fmt.Println("创建points错误:" + err.Error())
	}

	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     lf.server.Addr,
		Username: lf.server.User,
		Password: lf.server.Passwd,
	})
	c.Close()
	if err != nil {
		fmt.Println("创建客户端错误:" + err.Error())
	}
	lf.client = c
	lf.batchPoints = bp
	return lf
}
