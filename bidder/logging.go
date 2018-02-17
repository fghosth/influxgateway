package bidder

import (
	_ "fmt"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/influxdata/influxdb/client/v2"
	"jvole.com/influx/db"
)

var ()

type loggingService struct {
	logger log.Logger
	next   ServerBidder
}

func NewLoggingService(logger log.Logger, s ServerBidder) ServerBidder {
	return &loggingService{logger, s}
}
func (s *loggingService) Insert(tags map[string]string, fields map[string]interface{}, table string) error {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "Insert",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.next.Insert(tags, fields, table)
}
func (s *loggingService) InsertBatch(data []db.IndbData, table string) error {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "InsertBatch",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.next.InsertBatch(data, table)
}
func (s *loggingService) InsertNow(tags map[string]string, fields map[string]interface{}, table string) error {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "InsertNow",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.next.InsertNow(tags, fields, table)
}
func (s *loggingService) Query(cmd, db, precision string, limit, offset int) (res []client.Result, err error) {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "Query",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.next.Query(cmd, db, precision, limit, offset)
}
