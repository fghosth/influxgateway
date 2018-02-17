package route

import (
	_ "fmt"
	"time"

	"github.com/go-kit/kit/metrics"
	"jvole.com/influx/db"
)

type instrumentingService struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram

	next ServerRoute
}

func NewInstrumentingService(counter metrics.Counter, latency metrics.Histogram, s ServerRoute) ServerRoute {
	return &instrumentingService{
		requestCount:   counter,
		requestLatency: latency,
		next:           s,
	}
}
func (s *instrumentingService) Insert(tags map[string]string, fields map[string]interface{}, table string, uid uint64) error {
	defer func(begin time.Time) {
		s.requestCount.With("method", "Insert").Add(1)
		s.requestLatency.With("method", "Insert").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return s.next.Insert(tags, fields, table, uid)
}
func (s *instrumentingService) InsertNow(tags map[string]string, fields map[string]interface{}, table string, uid uint64) error {
	defer func(begin time.Time) {
		s.requestCount.With("method", "InsertNow").Add(1)
		s.requestLatency.With("method", "InsertNow").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return s.next.InsertNow(tags, fields, table, uid)
}
func (s *instrumentingService) InsertBatch(data []db.IndbData, table string, uid uint64) error {
	defer func(begin time.Time) {
		s.requestCount.With("method", "InsertBatch").Add(1)
		s.requestLatency.With("method", "InsertBatch").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return s.next.InsertBatch(data, table, uid)
}
func (s *instrumentingService) Select(cmd, db, precision string, limit, offset int, uid []uint64) (res []QueryResult, err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "Select").Add(1)
		s.requestLatency.With("method", "Select").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return s.next.Select(cmd, db, precision, limit, offset, uid)
}

func (s *instrumentingService) Query(cmd, db, precision string, uid []uint64) (res []QueryResult, err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "Query").Add(1)
		s.requestLatency.With("method", "Query").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return s.next.Query(cmd, db, precision, uid)
}

func (s *instrumentingService) Delete(cmd, db, precision string, uid uint64) (err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "Delete").Add(1)
		s.requestLatency.With("method", "Delete").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return s.next.Delete(cmd, db, precision, uid)
}
