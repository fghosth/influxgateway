package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/k0kubun/pp"
	"jvole.com/influx/util"
)

type Influxdb interface {
	/*
	   写influx数据库 有缓存
	   @parm tags 标签相当于属性
	   @parm fields 存储的字段集合，key value
	   @parm table 表明
	   @return error
	*/
	InsertWithBuffer(tags map[string]string, fields map[string]interface{}, table string) error
	/*
	   批量写influx数据库
	   @parm []IndbData 数据集合
	   @parm table 表明
	   @return error
	*/
	InsertBatch(data []IndbData, table string) error
	/*
	   立刻写入缓存中数据到influxdb
	   @return error
	*/
	InsertNow() error
	/*
	   写influx数据库
	   @parm tags 标签相当于属性
	   @parm fields 存储的字段集合，key value
	   @parm table 表明
	   @parm times 时间
	   @return error
	*/
	Insert(tags map[string]string, fields map[string]interface{}, table string, times time.Time) error
	/*
		   查询
		   @parm cmd 查询语句
		   @parm db 参数
		   @parm precision 精确度
			 @parm limit 每次查询显示条数限制
			 @param offse 下标
		   @return []client.Result 结果
		   @return error
	*/
	Select(cmd, db, precision string, limit, offset int) (res []client.Result, err error)
	/*
	   删除
	   @parm cmd 查询语句
	   @parm db 参数
	   @parm precision 精确度
	   @return error
	*/
	Delete(cmd, db, precision string) (err error)
	/*
	   执行任何语句
	   @parm cmd 查询语句
	   @parm db 参数
	   @parm precision 精确度
	   @return []client.Result 结果
	   @return error
	*/
	Query(cmd, db, precision string) (res []client.Result, err error)
	/*
		   检查状态
			 @param timeout 超时时间
			 @return time 响应时间
			 @return string
		   @return error
	*/

	Ping(timeout time.Duration) (time.Duration, string, error)
	/*
	   返回字符串
	   @return string
	*/
	ToString() string
	/*
	   关闭连接
	   @return error
	*/
	Close() error
}

type influxdb struct {
	addr        string //连接地址
	user        string //用户名
	passwd      string //密码
	database    string //数据库名
	precision   string //精确度
	client      client.Client
	buff        uint16             //当前缓存数量
	batchPoints client.BatchPoints //数据集合
	mtx         sync.RWMutex
}
type IndbData struct { //批量插入的结构
	Tags   map[string]string
	Fields map[string]interface{}
	Time   time.Time //插入时的时间
}

var (
	Buffer      = uint16(1000) //缓存，达到一定数量后写数据库，可提高效率。默认值：1000
	MaxLine     = 500          //查询时最多显示行数
	ERRNOSELECT = errors.New("只支持select操作")
	ERRNODELETE = errors.New("只支持delete操作")
)

/*
 检查状态
 @param timeout 超时时间
 @return time 响应时间
 @return string
 @return error
*/
func (idb *influxdb) Ping(timeout time.Duration) (time.Duration, string, error) {
	return idb.client.Ping(timeout)
}

/*
 返回字符串
 @return string
*/
func (idb *influxdb) ToString() string {
	str := fmt.Sprintf("地址:%s 数据库:%s ", idb.addr, idb.database)

	return string(str)
}

/*
 关闭连接
 @return error
*/
func (idb *influxdb) Close() error {
	return idb.client.Close()
}

/*
 批量写influx数据库
 @parm []IndbData 数据集合
 @parm table 表明
 @return error
*/
func (idb *influxdb) InsertBatch(data []IndbData, table string) error {

	// Create a new point batch
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  idb.database,
		Precision: idb.precision,
	})
	if err != nil {
		util.Log.WithFields(logrus.Fields{
			"name": "错误",
			"err":  err,
		}).Errorln("创建bp出错了")
		return err
	}

	for _, v := range data {
		pt, err := client.NewPoint(table, v.Tags, v.Fields, v.Time)
		if err != nil {
			util.Log.WithFields(logrus.Fields{
				"name": "错误",
				"err":  err,
			}).Errorln("循环出错了")
			return err
		}
		bp.AddPoint(pt)

		if uint16(len(bp.Points())) >= Buffer { //达到缓存上限写数据库
			err = writeNow(idb.client, bp, idb.addr)
			bp.ClearPoint()
			util.Log.WithFields(logrus.Fields{
				"name":    "批量插入",
				"package": "influx/db",
				"info":    "缓存满插入",
			}).Debugln("调试信息")
			if err != nil {
				return err
			}
		}
	}
	// pp.Println(len(bp.Points()))
	//写数据库
	// if err := idb.client.Write(bp); err != nil { //这里会把所有有效数据添加
	// 	util.Log.WithFields(logrus.Fields{
	// 		"name": "错误",
	// 		"err":  err,
	// 	}).Errorln("写出错了")
	// 	saveRecord(bp.Points(), idb.addr) //记录错误数据
	// 	return err
	// }
	if len(bp.Points()) > 0 {
		util.Log.WithFields(logrus.Fields{
			"name":    "批量插入",
			"package": "influx/db",
			"info":    "插入剩余数据",
		}).Debugln("调试信息")
		err = writeNow(idb.client, bp, idb.addr)
	}
	return err
}

//写入数据
func writeNow(cl client.Client, bp client.BatchPoints, addr string) error {
	var err error
	//写数据库
	if err = cl.Write(bp); err != nil { //这里会把所有有效数据添加
		util.Log.WithFields(logrus.Fields{
			"name": "错误",
			"err":  err,
		}).Errorln("写出错了")
		saveRecord(bp.Points(), addr) //记录错误数据
		return err
	}
	return err
}

/*
 查询
 @parm cmd 查询语句
 @parm db 参数
 @parm precision 精确度
 @return []client.Result 结果
 @return error
*/
func (idb *influxdb) Select(cmd, db, precision string, limit, offset int) (res []client.Result, err error) {
	if !strings.HasPrefix(strings.ToLower(cmd), "select") {
		err = ERRNOSELECT
		return
	}
	// Make client
	if limit > MaxLine { //最多显示多少行
		limit = MaxLine
	}
	cmd = cmd + " limit " + strconv.Itoa(limit) + " offset " + strconv.Itoa(offset)
	return idb.Query(cmd, db, precision)
}

/*
 删除
 @parm cmd 查询语句
 @parm db 参数
 @parm precision 精确度
 @return error
*/
func (idb *influxdb) Delete(cmd, db, precision string) (err error) {
	if !strings.HasPrefix(strings.ToLower(cmd), "delete") {
		err = ERRNODELETE
		return
	}
	_, err = idb.Query(cmd, db, precision)
	return err
}

/*
 执行任何语句
 @parm cmd 查询语句
 @parm db 参数
 @parm precision 精确度
 @return []client.Result 结果
 @return error
*/
func (idb *influxdb) Query(cmd, db, precision string) (res []client.Result, err error) {
	q := client.NewQuery(cmd, db, precision)
	if response, err := idb.client.Query(q); err == nil && response.Error() == nil {
		return response.Results, err
	} else {
		return nil, err
	}
}

/*
 立刻写入缓存中数据到influxdb
 @return error
*/
func (idb *influxdb) InsertNow() error {
	idb.mtx.Lock()
	defer idb.mtx.Unlock()
	//写数据库
	if err := idb.client.Write(idb.batchPoints); err != nil { //这里会把所有有效数据添加
		util.Log.WithFields(logrus.Fields{
			"name": "错误",
			"err":  err,
		}).Errorln("出错了")
		//TODO 保存数据用于回复
		saveRecord(idb.batchPoints.Points(), idb.addr)
		idb.batchPoints.ClearPoint() //源码增加了这个方法 注意包
		// pp.Println(idb.batchPoints)
		return err
	}
	idb.batchPoints.ClearPoint() //源码增加了这个方法 注意包
	idb.buff = 0
	return nil
}

//记录错误数据到本机
func saveRecord(point []*client.Point, addr string) {
	var data struct {
		Addr   string
		Tags   map[string]string
		Fields map[string]interface{}
		Table  string
		Time   time.Time
	}

	for _, v := range point {
		key := addr + "||" + util.GetGuid()
		data.Addr = addr
		data.Tags = v.Tags()
		data.Fields, _ = v.Fields()
		data.Time = v.Time()
		data.Table = v.Name()
		str, _ := json.Marshal(data)
		RecordE.Save(key, string(str))

		// pp.Println(RecordE.Load("https://localhost:8086||a1412f7706e21f76ed2a27cd912f52a9"))
		pp.Println(RecordE.Search("https", 4))
	}
}

/*
   写influx数据库
   @parm tags 标签相当于属性
   @parm fields 存储的字段集合，key value
   @parm precision  精度 h,m,s,ms,ns
   @parm table 表明
   @error
*/
func (idb *influxdb) InsertWithBuffer(tags map[string]string, fields map[string]interface{}, table string) error {

	pt, err := client.NewPoint(table, tags, fields, time.Now())
	if err != nil {
		util.Log.WithFields(logrus.Fields{
			"name": "错误",
			"err":  err,
		}).Errorln("出错了")
		return err
	}
	idb.mtx.Lock()
	idb.batchPoints.AddPoint(pt)
	idb.buff++
	idb.mtx.Unlock()
	if idb.buff >= Buffer { //缓存满了写数据库
		return idb.InsertNow()
	}
	return nil
}

/*
 写influx数据库
 @parm tags 标签相当于属性
 @parm fields 存储的字段集合，key value
 @parm table 表明
 @parm times 时间
 @return error
*/
func (idb *influxdb) Insert(tags map[string]string, fields map[string]interface{}, table string, times time.Time) error {
	// Create a new point batch
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  idb.database,
		Precision: idb.precision,
	})
	if err != nil {
		util.Log.WithFields(logrus.Fields{
			"name": "错误",
			"err":  err,
		}).Errorln("创建bp出错了")
		return err
	}

	pt, err := client.NewPoint(table, tags, fields, times)
	if err != nil {
		util.Log.WithFields(logrus.Fields{
			"name": "错误",
			"err":  err,
		}).Errorln("添加point出错了")
		return err
	}
	bp.AddPoint(pt)

	//写数据库
	if err := idb.client.Write(bp); err != nil { //这里会把所有有效数据添加
		util.Log.WithFields(logrus.Fields{
			"name": "错误",
			"err":  err,
		}).Errorln("写出错了")
		return err
	}
	return nil
}

func NewInfluxdb(addr, user, password, db, precision string) *influxdb {
	// fmt.Printf("addr:%s,user:%s,pwd:%s\n", addr, user, password)
	// 创建 point batch
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  db,
		Precision: precision,
	})
	if err != nil {
		util.Log.WithFields(logrus.Fields{
			"name": "错误",
			"err":  err,
		}).Errorln("出错了")
	}

	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     addr,
		Username: user,
		Password: password,
	})
	if err != nil {
		util.Log.WithFields(logrus.Fields{
			"name": "错误",
			"err":  err,
		}).Errorln("出错了")
	}
	indb := &influxdb{addr, user, password, db, precision, c, 0, bp, *new(sync.RWMutex)}
	return indb
}
