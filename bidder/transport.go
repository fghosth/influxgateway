package bidder

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"jvole.com/influx/db"

	kitlog "github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"
)

var errBadRoute = errors.New("bad route")

func MakeHandler(bs ServerBidder, logger kitlog.Logger) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorLogger(logger),
		kithttp.ServerErrorEncoder(encodeError),
	}
	r := mux.NewRouter()
	InsertHandler := kithttp.NewServer(
		makeInsertEndpoint(bs),
		decodeInsertRequest,
		encodeResponse,
		opts...,
	)
	QueryHandler := kithttp.NewServer(
		makeQueryEndpoint(bs),
		decodeQueryRequest,
		encodeResponse,
		opts...,
	)
	InsertNowHandler := kithttp.NewServer(
		makeInsertNowEndpoint(bs),
		decodeInsertNowRequest,
		encodeResponse,
		opts...,
	)
	InsertBatchHandler := kithttp.NewServer(
		makeInsertBatchEndpoint(bs),
		decodeInsertBatchRequest,
		encodeResponse,
		opts...,
	)
	r.Handle("/influx/v1/insert", InsertHandler).Methods("POST")
	r.Handle("/influx/v1/query", QueryHandler).Methods("POST")
	r.Handle("/influx/v1/insertNow", InsertNowHandler).Methods("POST")
	r.Handle("/influx/v1/insertBatch", InsertBatchHandler).Methods("POST")
	return r
}
func decodeQueryRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var body struct {
		Cmd       string `json:"cmd"`
		Db        string `json:"db"`
		Precision string `json:"precision"`
		Limit     int    `json:"limit"`
		Offset    int    `json:"offset"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}

	// pp.Println(body)
	return queryRequest{
		body.Cmd,
		body.Db,
		body.Precision,
		body.Limit,
		body.Offset,
	}, nil
}

func decodeInsertBatchRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var body struct {
		Data  []db.IndbData `json:"data"`
		Table string        `json:"table"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}
	return insertBatchRequest{
		body.Data,
		body.Table,
	}, nil
}

func decodeInsertRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var body struct {
		Tags  map[string]string      `json:"tags"`
		Field map[string]interface{} `json:"field"`
		Table string                 `json:"table"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}
	return insertRequest{
		body.Tags,
		body.Field,
		body.Table,
	}, nil
}
func decodeInsertNowRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var body struct {
		Tags  map[string]string      `json:"tags"`
		Field map[string]interface{} `json:"field"`
		Table string                 `json:"table"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}

	return insertRequest{
		body.Tags,
		body.Field,
		body.Table,
	}, nil
}
func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
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
		"error": err.Error(),
	})
}
