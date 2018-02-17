package bidder

import (
	_ "fmt"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/influxdata/influxdb/client/v2"
	"jvole.com/influx/db"
)

var ()

type instrumentingService struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram

	next ServerBidder
}

func NewInstrumentingService(counter metrics.Counter, latency metrics.Histogram, s ServerBidder) ServerBidder {
	return &instrumentingService{
		requestCount:   counter,
		requestLatency: latency,
		next:           s,
	}
}
func (s *instrumentingService) Insert(tags map[string]string, fields map[string]interface{}, table string) error {
	defer func(begin time.Time) {
		s.requestCount.With("method", "insert").Add(1)
		s.requestLatency.With("method", "insert").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return s.next.Insert(tags, fields, table)
}
func (s *instrumentingService) InsertBatch(data []db.IndbData, table string) error {
	defer func(begin time.Time) {
		s.requestCount.With("method", "InsertBatch").Add(1)
		s.requestLatency.With("method", "InsertBatch").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return s.next.InsertBatch(data, table)
}
func (s *instrumentingService) InsertNow(tags map[string]string, fields map[string]interface{}, table string) error {
	defer func(begin time.Time) {
		s.requestCount.With("method", "InsertNow").Add(1)
		s.requestLatency.With("method", "InsertNow").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return s.next.InsertNow(tags, fields, table)
}

func (s *instrumentingService) Query(cmd, db, precision string, limit, offset int) (res []client.Result, err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "query").Add(1)
		s.requestLatency.With("method", "query").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return s.next.Query(cmd, db, precision, limit, offset)
}
