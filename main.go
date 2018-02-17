package main

import (
	glog "log"
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/fsnotify/fsnotify"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/go-kit/kit/log"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"jvole.com/influx/bidder"
	"jvole.com/influx/db"
	"jvole.com/influx/util"
)

var cfg *config

type config struct {
	port      string //端口号 :8000
	addr      string //infulx地址
	username  string //influx用户名
	password  string //influx密码
	dbname    string //influx数据库名
	precision string //精确度
	buffer    uint16 //缓存
}

func init() {
	util.Viper.SetConfigType("yaml")
	// util.Viper.SetConfigName(".cfg")
	util.Viper.AddConfigPath(".")
	// util.Viper.AddConfigPath("/Users/derek/project/go/src/jvole.com/monitor/")
	util.Viper.SetConfigFile("server.yaml")

	err := util.Viper.ReadInConfig() // Find and read the config file
	if err != nil {                  // Handle errors reading the config file
		util.Log.WithFields(logrus.Fields{
			"name": "信息",
			"err":  err,
		}).Infoln("配置文件加载失败")
		return
	}

	cfg = &config{}
	mapConfig()
	// pp.Println(cfg.buffer)
	// db.Buffer = cfg.buffer
	//监控配置文件变化
	util.Viper.WatchConfig()
	util.Viper.OnConfigChange(func(e fsnotify.Event) {
		util.Log.WithFields(logrus.Fields{
			"name": "信息",
		}).Infoln("配置文件变更，重新生效")
		mapConfig()
		db.Buffer = cfg.buffer
		// pp.Println(cfg.buffer)
	})
}
func main() {
	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	fieldKeys := []string{"method"}
	var bs bidder.ServerBidder
	bs = bidder.NewServer(cfg.addr, cfg.username, cfg.password, cfg.dbname, cfg.precision, cfg.buffer)
	bs = bidder.NewLoggingService(log.With(logger, "component", "influxdb"), bs)
	bs = bidder.NewInstrumentingService(
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "api",
			Subsystem: "influx_service",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, fieldKeys),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "api",
			Subsystem: "influx_service",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, fieldKeys),
		bs)

	httpLogger := log.With(logger, "component", "http")

	mux := http.NewServeMux()

	mux.Handle("/influx/v1/", bidder.MakeHandler(bs, httpLogger))
	http.Handle("/", accessControl(mux))
	http.Handle("/metrics", promhttp.Handler())
	glog.Fatal(http.ListenAndServe(cfg.port, nil))
}

func accessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}

func mapConfig() {
	if influxcfg, ok := util.Viper.Get("inflxudb").(map[string]interface{}); ok {
		cfg.addr = influxcfg["addr"].(string)
		cfg.dbname = influxcfg["dbname"].(string)
		cfg.precision = influxcfg["precision"].(string)
		cfg.username = influxcfg["username"].(string)
		cfg.password = influxcfg["password"].(string)
	}
	if v, ok := util.Viper.Get("buffer").(int); ok {

		cfg.buffer = uint16(v)
	}
	if v, ok := util.Viper.Get("port").(string); ok {
		cfg.port = v
	}

}
