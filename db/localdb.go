package db

type RecordErrData interface {
	/*
	   记录
	   @parm key
	   @parm value
	   @return error
	*/
	Save(key, value string) error
	/*
	   查询key
	   @parm key
	   @return string
	*/
	Load(key string) string
	/*
	   删除数据
	   @parm key
	   @return err
	*/
	Delete(key string) error
	/*
	   根据前缀查询
	   @parm prefix 前缀
	   @return []结果集
	*/
	Search(prefix string, num int) []string
}

var RecordE RecordErrData
