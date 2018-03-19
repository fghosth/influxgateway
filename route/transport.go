package route

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"path"
	"runtime"
	"time"

	"github.com/gorilla/mux"
	// "golang.org/x/time/rate"
	"github.com/go-kit/kit/auth/basic"
	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	kithttp "github.com/go-kit/kit/transport/http"
	"jvole.com/influx/db"
	"jvole.com/influx/middleware"
	"jvole.com/influx/util"
)

var errBadRoute = errors.New("bad route")
var User, Password string
var qps = 100000

func MakeHandler(bs ServerRoute, logger *kitlog.Logger) http.Handler {

	opts := []kithttp.ServerOption{
		kithttp.ServerBefore(kithttp.PopulateRequestContext),
		kithttp.ServerErrorLogger(*logger),
		kithttp.ServerErrorEncoder(encodeError),
	}
	r := mux.NewRouter()
	e := makeInsertNowEndpoint(bs)
	e = basic.AuthMiddleware(User, Password, "")(e)
	e = middleware.ValidMiddleware()(e)
	// e = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), qps))(e)
	InsertNowHandler := kithttp.NewServer(
		e,
		decodeInsertNowRequest,
		encodeResponse,
		opts...,
	)
	eib := makeInsertBatchEndpoint(bs)
	eib = basic.AuthMiddleware(User, Password, "")(eib)
	eib = middleware.ValidMiddleware()(eib)
	// eib = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), qps))(eib)
	InsertBatchHandler := kithttp.NewServer(
		eib,
		decodeInsertBatchRequest,
		encodeResponse,
		opts...,
	)
	es := makeSelectEndpoint(bs)
	es = basic.AuthMiddleware(User, Password, "")(es)
	es = middleware.ValidMiddleware()(es)
	// es = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), qps))(es)
	SelectHandler := kithttp.NewServer(
		es,
		decodeSelectRequest,
		encodeResponse,
		opts...,
	)
	ei := makeInsertEndpoint(bs)
	ei = basic.AuthMiddleware(User, Password, "")(ei)
	ei = middleware.ValidMiddleware()(ei)
	// ei = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), qps))(ei)
	InsertHandler := kithttp.NewServer(
		ei,
		decodeInsertRequest,
		encodeResponse,
		opts...,
	)
	ed := makeDeleteEndpoint(bs)
	ed = basic.AuthMiddleware(User, Password, "")(ed)
	ed = middleware.ValidMiddleware()(ed)
	// ed = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), qps))(ed)
	DeleteHandler := kithttp.NewServer(
		ed,
		decodeDeleteRequest,
		encodeResponse,
		opts...,
	)
	eq := makeQueryEndpoint(bs)
	eq = basic.AuthMiddleware("fghosth", "zaq1xsw2CDE#", "")(eq)
	eq = middleware.ValidMiddleware()(eq)
	// eq = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), qps))(eq)
	QueryHandler := kithttp.NewServer(
		eq,
		decodeQueryRequest,
		encodeResponse,
		opts...,
	)
	r.Handle("/influx/v2/insertBatch", InsertBatchHandler).Methods("POST")
	r.Handle("/influx/v2/select", SelectHandler).Methods("POST")
	r.Handle("/influx/v2/insert", InsertHandler).Methods("POST")
	r.Handle("/influx/v2/insertNow", InsertNowHandler).Methods("POST")
	r.Handle("/influx/v2/delete", DeleteHandler).Methods("POST")
	r.Handle("/influx/v2/query", QueryHandler).Methods("POST")
	return r
}
func decodeInsertRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var body struct {
		UID   uint64                 `json:"uid"`
		Tags  map[string]string      `json:"tags"`
		Field map[string]interface{} `json:"field"`
		Table string                 `json:"table"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}
	defer func(begin time.Time) {
		pc, file, line, _ := runtime.Caller(1)
		f := runtime.FuncForPC(pc)
		level.Debug(util.KitLogger).Log(
			"method", f.Name(),
			"file", path.Base(file),
			"line", line,
			"request", body,
			"took", time.Since(begin).Nanoseconds()/1000,
		)
	}(time.Now())
	return InsertRequest{
		body.UID,
		body.Tags,
		body.Field,
		body.Table,
	}, nil
}
func decodeInsertNowRequest(_ context.Context, r *http.Request) (interface{}, error) {

	var body struct {
		UID   uint64                 `json:"uid"`
		Tags  map[string]string      `json:"tags"`
		Field map[string]interface{} `json:"field"`
		Table string                 `json:"table"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}
	defer func(begin time.Time) {
		pc, file, line, _ := runtime.Caller(1)
		f := runtime.FuncForPC(pc)
		level.Debug(util.KitLogger).Log(
			"method", f.Name(),
			"file", path.Base(file),
			"line", line,
			"request", body,
			"took", time.Since(begin).Nanoseconds()/1000,
		)
	}(time.Now())
	return InsertRequest{
		body.UID,
		body.Tags,
		body.Field,
		body.Table,
	}, nil
}
func decodeInsertBatchRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var body struct {
		UID   uint64        `json:"uid"`
		Data  []db.IndbData `json:"data"`
		Table string        `json:"table"`
	}
	// a, _ := time.Parse("2018-02-03 15:04:05", "2018-02-03 15:04:05")
	// a := time.NewTimer(time.Duration(body.Time))
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}
	defer func(begin time.Time) {
		pc, file, line, _ := runtime.Caller(1)
		f := runtime.FuncForPC(pc)
		level.Debug(util.KitLogger).Log(
			"method", f.Name(),
			"file", path.Base(file),
			"line", line,
			"request", body,
			"took", time.Since(begin).Nanoseconds()/1000,
		)
	}(time.Now())
	return InsertBatchRequest{
		body.UID,
		body.Data,
		body.Table,
	}, nil
}
func decodeSelectRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var body struct {
		UID       []uint64 `json:"uid"`
		Cmd       string   `json:"cmd"`
		Db        string   `json:"db"`
		Precision string   `json:"precision"`
		Limit     int      `json:"limit"`
		Offset    int      `json:"offset"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}
	defer func(begin time.Time) {
		pc, file, line, _ := runtime.Caller(1)
		f := runtime.FuncForPC(pc)
		level.Debug(util.KitLogger).Log(
			"method", f.Name(),
			"file", path.Base(file),
			"line", line,
			"request", body,
			"took", time.Since(begin).Nanoseconds()/1000,
		)
	}(time.Now())
	// pp.Println(body)
	return SelectRequest{
		body.UID,
		body.Cmd,
		body.Db,
		body.Precision,
		body.Limit,
		body.Offset,
	}, nil
}

func decodeQueryRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var body struct {
		UID       []uint64 `json:"uid"`
		Cmd       string   `json:"cmd"`
		Db        string   `json:"db"`
		Precision string   `json:"precision"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}
	defer func(begin time.Time) {
		pc, file, line, _ := runtime.Caller(1)
		f := runtime.FuncForPC(pc)
		level.Debug(util.KitLogger).Log(
			"method", f.Name(),
			"file", path.Base(file),
			"line", line,
			"request", body,
			"took", time.Since(begin).Nanoseconds()/1000,
		)
	}(time.Now())
	// pp.Println(body)
	return QueryRequest{
		body.UID,
		body.Cmd,
		body.Db,
		body.Precision,
	}, nil
}

func decodeDeleteRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var body struct {
		UID       uint64 `json:"uid"`
		Cmd       string `json:"cmd"`
		Db        string `json:"db"`
		Precision string `json:"precision"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}
	defer func(begin time.Time) {
		pc, file, line, _ := runtime.Caller(1)
		f := runtime.FuncForPC(pc)
		level.Debug(util.KitLogger).Log(
			"method", f.Name(),
			"file", path.Base(file),
			"line", line,
			"request", body,
			"took", time.Since(begin).Nanoseconds()/1000,
		)
	}(time.Now())
	// pp.Println(body)
	return DeleteRequest{
		body.UID,
		body.Cmd,
		body.Db,
		body.Precision,
	}, nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	defer func(begin time.Time) {
		pc, file, line, _ := runtime.Caller(1)
		f := runtime.FuncForPC(pc)
		level.Debug(util.KitLogger).Log(
			"method", f.Name(),
			"file", path.Base(file),
			"line", line,
			"response", response,
			"took", time.Since(begin).Nanoseconds()/1000,
		)
	}(time.Now())
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

type errorer interface {
	error() error
}

// encode errors from business-logic
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch err {
	case errBadRoute:
		w.WriteHeader(http.StatusNotFound)
	case errBadRoute:
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"errcode": "1003",
		"msg":     err.Error(),
		"data":    nil,
	})
}
