package bidder

import (
	"context"
	_ "fmt"

	"github.com/go-kit/kit/endpoint"
	"jvole.com/influx/db"
)

var ()

type insertBatchRequest struct {
	Data  []db.IndbData
	Table string
}
type insertBatchResponse struct {
	Errcode string
	Msg     string
	Data    map[string]string
	Err     error `json:"error,omitempty"`
}

func (r insertBatchResponse) error() error { return r.Err }
func makeInsertBatchEndpoint(s ServerBidder) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(insertBatchRequest)
		err := s.InsertBatch(req.Data, req.Table)
		if err != nil {
			// pp.Println(err)
			return insertBatchResponse{Errcode: "00003", Msg: "insert error", Data: nil, Err: err}, nil
		}
		return insertBatchResponse{Errcode: "0", Msg: "ok", Data: nil, Err: err}, nil
	}
}

type insertRequest struct {
	Tags  map[string]string
	Field map[string]interface{}
	Table string
}
type insertResponse struct {
	Errcode string
	Msg     string
	Data    map[string]string
	Err     error `json:"error,omitempty"`
}

func (r insertResponse) error() error { return r.Err }
func makeInsertEndpoint(s ServerBidder) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(insertRequest)
		err := s.Insert(req.Tags, req.Field, req.Table)
		if err != nil {
			// pp.Println(err)
			return insertResponse{Errcode: "00002", Msg: "insert error", Data: nil, Err: err}, nil
		}
		return insertResponse{Errcode: "0", Msg: "ok", Data: nil, Err: err}, nil
	}
}

func makeInsertNowEndpoint(s ServerBidder) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(insertRequest)
		err := s.InsertNow(req.Tags, req.Field, req.Table)
		if err != nil {
			// pp.Println(err)
			return insertResponse{Errcode: "00002", Msg: "insert error", Data: nil, Err: err}, nil
		}
		return insertResponse{Errcode: "0", Msg: "ok", Data: nil, Err: err}, nil
	}
}

type queryRequest struct {
	Cmd       string
	Db        string
	Precision string
	Limit     int
	Offset    int
}

type queryResponse struct {
	Errcode string
	Msg     string
	Data    interface{}
	Err     error `json:"error,omitempty"`
}

func (r queryResponse) error() error { return r.Err }
func makeQueryEndpoint(s ServerBidder) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {

		req := request.(queryRequest)
		result, err := s.Query(req.Cmd, req.Db, req.Precision, req.Limit, req.Offset)
		if err != nil {
			return queryResponse{Errcode: "00001", Msg: "error", Data: nil, Err: err}, nil
		}
		var res interface{}
		if result != nil {
			if result[0].Series != nil {
				res = result[0].Series[0].Values
			}
		}
		return queryResponse{Errcode: "0", Msg: "ok", Data: res, Err: err}, nil
	}
}
