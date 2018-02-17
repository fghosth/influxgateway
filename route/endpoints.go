package route

import (
	"context"
	_ "fmt"
	"strings"

	"github.com/go-kit/kit/endpoint"
	"jvole.com/influx/db"
)

func makeInsertNowEndpoint(s ServerRoute) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(InsertRequest)
		if req.Uid == 0 {
			// pp.Println(err)
			return InsertBatchResponse{Errcode: "00004", Msg: "Uid不能为0", Data: nil, Err: nil}, nil
		}
		err := s.InsertNow(req.Tags, req.Field, req.Table, req.Uid)
		if err != nil {
			// pp.Println(err)
			return InsertResponse{Errcode: "00002", Msg: "insert error", Data: nil, Err: err}, nil
		}
		return InsertResponse{Errcode: "0", Msg: "ok", Data: nil, Err: err}, nil
	}
}

type InsertBatchRequest struct {
	Uid   uint64        `json:"uid" valid:"required~用户id不能为空"`
	Data  []db.IndbData `json:"data" valid:"-"`
	Table string        `json:"table" valid:"required~表名不能为空"`
}
type InsertBatchResponse struct {
	Errcode string
	Msg     string
	Data    map[string]string
	Err     error `json:"error,omitempty"`
}

func (r InsertBatchResponse) error() error { return r.Err }
func makeInsertBatchEndpoint(s ServerRoute) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(InsertBatchRequest)
		if req.Uid == 0 {
			// pp.Println(err)
			return InsertBatchResponse{Errcode: "00004", Msg: "Uid不能为0", Data: nil, Err: nil}, nil
		}
		err := s.InsertBatch(req.Data, req.Table, req.Uid)
		if err != nil {
			// pp.Println(err)
			return InsertBatchResponse{Errcode: "00003", Msg: "insert error", Data: nil, Err: err}, nil
		}
		return InsertBatchResponse{Errcode: "0", Msg: "ok", Data: nil, Err: err}, nil
	}
}

type SelectRequest struct {
	Uid       []uint64 `json:"uid" valid:"uidarr~至少需要一个uid"`
	Cmd       string   `json:"cmd" valid:"required~命令不cmd能为空,matches(^select.*)~只能执行select操作"`
	Db        string   `json:"db" valid:"required~数据库名db不能为空"`
	Precision string   `json:"precision" valid:"required~精度limit不能为空,in(d|h|m|s|ms|ns)~必须为(d|h|m|s|ms|ns)中的一个"`
	Limit     int      `json:"limit" valid:"required~limit不能为空"`
	Offset    int      `json:"offset" valid:"-"`
}

type SelectResponse struct {
	Errcode string
	Msg     string
	Data    interface{}
	Err     error `json:"error,omitempty"`
}

func (r SelectResponse) error() error { return r.Err }

func makeSelectEndpoint(s ServerRoute) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(SelectRequest)
		result, err := s.Select(req.Cmd, req.Db, req.Precision, req.Limit, req.Offset, req.Uid)
		if !strings.HasPrefix(strings.ToLower(req.Cmd), "select") {
			return SelectResponse{Errcode: "00005", Msg: "只能支持Select操作", Data: nil, Err: err}, nil
		}
		data := make(map[uint64]interface{})
		for _, v := range result { //输出简洁的json格式
			if len(v.Result) >= 1 {
				if len(v.Result[0].Series) >= 1 {
					data[v.Uid] = v.Result[0].Series
				}
			} else {
				data[v.Uid] = nil
			}
		}
		if err != nil {
			return SelectResponse{Errcode: "00001", Msg: "error", Data: nil, Err: err}, nil
		}
		return SelectResponse{Errcode: "0", Msg: "ok", Data: data, Err: err}, nil
	}
}

type InsertRequest struct {
	Uid   uint64                 `json:"uid" valid:"required~用户id不能为空"`
	Tags  map[string]string      `valid:"-"`
	Field map[string]interface{} `valid:"-"`
	Table string                 `json:"table" valid:"required~表名不能为空"`
}
type InsertResponse struct {
	Errcode string
	Msg     string
	Data    map[string]string
	Err     error `json:"error,omitempty"`
}

func (r InsertResponse) error() error { return r.Err }
func makeInsertEndpoint(s ServerRoute) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(InsertRequest)
		if req.Uid == 0 {
			// pp.Println(err)
			return InsertBatchResponse{Errcode: "00004", Msg: "Uid不能为0", Data: nil, Err: nil}, nil
		}
		err := s.Insert(req.Tags, req.Field, req.Table, req.Uid)
		if err != nil {
			// pp.Println(err)
			return InsertResponse{Errcode: "00002", Msg: "insert error", Data: nil, Err: err}, nil
		}
		return InsertResponse{Errcode: "0", Msg: "ok", Data: nil, Err: err}, nil
	}
}

type DeleteRequest struct {
	Uid       uint64 `json:"uid" valid:"required~用户id不能为空"`
	Cmd       string `json:"cmd" valid:"required~命令不cmd能为空,matches(^delete.*)~只能执行delete操作"`
	Db        string `json:"db" valid:"required~数据库名db不能为空"`
	Precision string `json:"precision" valid:"required~精度limit不能为空,in(d|h|m|s|ms|ns)~必须为(d|h|m|s|ms|ns)中的一个"`
}
type DeleteResponse struct {
	Errcode string
	Msg     string
	Data    map[string]string
	Err     error `json:"error,omitempty"`
}

func (r DeleteResponse) error() error { return r.Err }
func makeDeleteEndpoint(s ServerRoute) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(DeleteRequest)
		if req.Uid == 0 {
			// pp.Println(err)
			return InsertBatchResponse{Errcode: "00004", Msg: "Uid不能为0", Data: nil, Err: nil}, nil
		}
		if !strings.HasPrefix(strings.ToLower(req.Cmd), "delete") {
			return DeleteResponse{Errcode: "00007", Msg: "只能支持Delete操作", Data: nil, Err: nil}, nil
		}
		err := s.Delete(req.Cmd, req.Db, req.Precision, req.Uid)
		if err != nil {
			// pp.Println(err)
			return InsertResponse{Errcode: "00006", Msg: "Delete error", Data: nil, Err: err}, nil
		}
		return InsertResponse{Errcode: "0", Msg: "ok", Data: nil, Err: err}, nil
	}
}

type QueryRequest struct {
	Uid       []uint64 `json:"uid" valid:"-"`
	Cmd       string   `json:"cmd" valid:"required~命令不cmd能为空"`
	Db        string   `json:"db" valid:"required~数据库名db不能为空"`
	Precision string   `json:"precision" valid:"required~精度limit不能为空,in(d|h|m|s|ms|ns)~必须为(d|h|m|s|ms|ns)中的一个"`
}

type QueryResponse struct {
	Errcode string
	Msg     string
	Data    interface{}
	Err     error `json:"error,omitempty"`
}

func (r QueryResponse) error() error { return r.Err }

func makeQueryEndpoint(s ServerRoute) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(QueryRequest)
		result, err := s.Query(req.Cmd, req.Db, req.Precision, req.Uid)
		data := make(map[uint64]interface{})
		for _, v := range result { //输出简洁的json格式
			if len(v.Result) > 0 {
				// pp.Println(v.Result[0].Series)
				if len(v.Result[0].Series) > 0 {
					data[v.Uid] = v.Result[0].Series
				}
			} else {
				data[v.Uid] = nil
			}
		}
		if err != nil {
			return QueryResponse{Errcode: "00001", Msg: "error", Data: nil, Err: err}, nil
		}
		return QueryResponse{Errcode: "0", Msg: "ok", Data: data, Err: err}, nil
	}
}
