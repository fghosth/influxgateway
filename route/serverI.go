package route

import (
	"github.com/influxdata/influxdb/client/v2"
	"jvole.com/influx/db"
)

type ServerRoute interface {
	/*
	   写influx数据库 有缓存满足数量才真的写数据库
	   @parm tags 标签相当于属性
	   @parm fields 存储的字段集合，key value
	   @parm table 表明
	   @parm uid 某用户的id 根据用户id存放数据，某一用户的数据始终存放在同一台server上
	   @return error
	*/
	Insert(tags map[string]string, fields map[string]interface{}, table string, uid uint64) error
	/*
	 写influx数据库 无缓存立刻写数据库
	 @parm tags 标签相当于属性
	 @parm fields 存储的字段集合，key value
	 @parm table 表明
	   @parm uid 某用户的id 根据用户id存放数据，某一用户的数据始终存放在同一台server上
	 @return error
	*/
	InsertNow(tags map[string]string, fields map[string]interface{}, table string, uid uint64) error
	/*
	 批量写influx数据库 立刻写数据
	 @parm tags 标签相当于属性
	 @parm fields 存储的字段集合，key value
	 @parm table 表明
	   @parm uid 某用户的id 根据用户id存放数据，某一用户的数据始终存放在同一台server上
	 @return error
	*/
	InsertBatch(data []db.IndbData, table string, uid uint64) error
	/*
	   查询
	   @parm cmd 查询语句
	   @parm db 参数
	   @parm precision 精确度
	   @parm uid 某用户的id数组 根据用户id查询，某一用户的数据始终存放在同一台server上
	   @return []client.QueryResult 结果
	   @return error
	*/
	Select(cmd, db, precision string, limit, offset int, uid []uint64) (res []QueryResult, err error)
	/*
	   删除
	   @parm cmd 删除语句
	   @parm db 参数
	   @parm precision 精确度
	   @parm uid 某用户的id 根据用户id查询，某一用户的数据始终存放在同一台server上
	   @return error
	*/
	Delete(cmd, db, precision string, uid uint64) (err error)
	/*
	   执行任何命令
	   @parm cmd 命令
	   @parm db 参数
	   @parm precision 精确度
	   @parm uid 某用户的id 根据用户id查询，某一用户的数据始终存放在同一台server上
	   @return []client.Result 结果
	   @return error
	*/
	Query(cmd, db, precision string, uid []uint64) (res []QueryResult, err error)
}
type QueryResult struct {
	Uid    uint64
	Result []client.Result
}
