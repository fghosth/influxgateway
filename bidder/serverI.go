package bidder

import (
	"github.com/influxdata/influxdb/client/v2"
	"jvole.com/influx/db"
)

type ServerBidder interface {
	/*
	   写influx数据库 有缓存满足数量才真的写数据库
	   @parm tags 标签相当于属性
	   @parm fields 存储的字段集合，key value
	   @parm table 表明
	   @return error
	*/
	Insert(tags map[string]string, fields map[string]interface{}, table string) error
	/*
	 写influx数据库 无缓存立刻写数据库
	 @parm tags 标签相当于属性
	 @parm fields 存储的字段集合，key value
	 @parm table 表明
	 @return error
	*/
	InsertNow(tags map[string]string, fields map[string]interface{}, table string) error
	/*
	 批量写influx数据库 立刻写数据
	 @parm tags 标签相当于属性
	 @parm fields 存储的字段集合，key value
	 @parm table 表明
	 @return error
	*/
	InsertBatch(data []db.IndbData, table string) error
	/*
	   查询
	   @parm cmd 查询语句
	   @parm db 参数
	   @parm precision 精确度
	   @return []client.Result 结果
	   @return error
	*/
	Query(cmd, db, precision string, limit, offset int) (res []client.Result, err error)
}
