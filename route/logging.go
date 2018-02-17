package route

import (
	_ "fmt"
	"path"
	"runtime"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"jvole.com/influx/db"
)

type loggingService struct {
	logger log.Logger
	next   ServerRoute
}

func NewLoggingService(logger log.Logger, s ServerRoute) ServerRoute {
	return &loggingService{logger, s}
}

func (s *loggingService) Select(cmd, db, precision string, limit, offset int, uid []uint64) (res []QueryResult, err error) {
	defer func(begin time.Time) {
		pc, file, line, _ := runtime.Caller(1)
		f := runtime.FuncForPC(pc)
		level.Info(s.logger).Log(
			"method", f.Name(),
			"file", path.Base(file),
			"line", line,
			"took", time.Since(begin).Nanoseconds()/1000,
		)
	}(time.Now())

	return s.next.Select(cmd, db, precision, limit, offset, uid)
}

func (s *loggingService) Query(cmd, db, precision string, uid []uint64) (res []QueryResult, err error) {
	defer func(begin time.Time) {
		pc, file, line, _ := runtime.Caller(1)
		f := runtime.FuncForPC(pc)
		level.Info(s.logger).Log(
			"method", f.Name(),
			"file", path.Base(file),
			"line", line,
			"took", time.Since(begin).Nanoseconds()/1000,
		)
	}(time.Now())
	return s.next.Query(cmd, db, precision, uid)
}

func (s *loggingService) Delete(cmd, db, precision string, uid uint64) (err error) {
	defer func(begin time.Time) {
		pc, file, line, _ := runtime.Caller(1)
		f := runtime.FuncForPC(pc)
		level.Info(s.logger).Log(
			"method", f.Name(),
			"file", path.Base(file),
			"line", line,
			"took", time.Since(begin).Nanoseconds()/1000,
		)
	}(time.Now())
	return s.next.Delete(cmd, db, precision, uid)
}

func (s *loggingService) Insert(tags map[string]string, fields map[string]interface{}, table string, uid uint64) error {
	defer func(begin time.Time) {
		pc, file, line, _ := runtime.Caller(1)
		f := runtime.FuncForPC(pc)
		level.Info(s.logger).Log(
			"method", f.Name(),
			"file", path.Base(file),
			"line", line,
			"took", time.Since(begin).Nanoseconds()/1000,
		)
	}(time.Now())
	return s.next.Insert(tags, fields, table, uid)
}
func (s *loggingService) InsertNow(tags map[string]string, fields map[string]interface{}, table string, uid uint64) error {
	defer func(begin time.Time) {
		pc, file, line, _ := runtime.Caller(1)
		f := runtime.FuncForPC(pc)
		level.Info(s.logger).Log(
			"method", f.Name(),
			"file", path.Base(file),
			"line", line,
			"took", time.Since(begin).Nanoseconds()/1000,
		)
	}(time.Now())
	return s.next.InsertNow(tags, fields, table, uid)
}
func (s *loggingService) InsertBatch(data []db.IndbData, table string, uid uint64) error {
	defer func(begin time.Time) {
		pc, file, line, _ := runtime.Caller(1)
		f := runtime.FuncForPC(pc)
		level.Info(s.logger).Log(
			"method", f.Name(),
			"file", path.Base(file),
			"line", line,
			"took", time.Since(begin).Nanoseconds()/1000,
		)
	}(time.Now())
	return s.next.InsertBatch(data, table, uid)
}
