package route

import (
	"errors"
	"time"

	"github.com/Sirupsen/logrus"
	"jvole.com/influx/db"
	"jvole.com/influx/util"
	"jvole.com/influx/weighted"
)

type InfluxGateway interface {
	/*检查配置组的正确性 必须是2的n次方
	  @return  bool true  false
	*/
	CheckConfig() bool
	/*根据id分组，取余hash算法。id不能重复
	  @Group  返回分组
	*/
	GetGroup(id uint64) (gn uint64, gp Group)
}
type Server struct {
	Addr       string //infulx地址
	Username   string //influx用户名
	Password   string //influx密码
	Dbname     string //influx数据库名
	Precision  string // 精确度
	Weight     int    //权重
	QueryCount uint64 //查询次数
	Status     bool   //服务器是否可用true 可用  FALSE 不可用
	Conn       db.Influxdb
}

type Group struct {
	Poll weighted.W
	Num  uint64
	S    []Server
}

var (
	Groups []Group //所有组
	Total  uint64  //所有的组数量
)

/*检查配置组的正确性 必须是2的n次方
  @return  bool true  false
*/
func (g Group) CheckConfig() bool {
	n := len(Groups)
	if !util.Is2N(n) { //如果不是2的N次方出错
		util.Log.WithFields(logrus.Fields{
			"name": "错误",
			"err":  errors.New("组数量必须是2的N次方"),
		}).Errorln("组数量必须是2的N次方")
		return false
	}
	for _, gv := range Groups { //遍历组
		for _, sv := range gv.S { //遍历server 服务不通报错
			_, _, err := sv.Conn.Ping(time.Second * 2) //检查服务器状态
			if err != nil {
				util.Log.WithFields(logrus.Fields{
					"name":  "错误",
					"err":   err,
					"group": gv.Num,
				}).Errorln("服务器连接失败")
				return false
			}
		}
	}
	return true
}

/*根据id分组，取余hash算法。id不能重复
  @Group  返回分组
*/
func (g Group) GetGroup(id uint64) (gn uint64, gp Group) {
	gn = id%Total + 1
	for _, v := range Groups {
		if v.Num == gn {
			gp = v
		}
	}
	return
}
