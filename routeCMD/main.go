package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/fsnotify/fsnotify"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/k0kubun/pp"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/net/netutil"
	"jvole.com/influx/db"
	"jvole.com/influx/loginflux"
	"jvole.com/influx/recover"
	"jvole.com/influx/route"
	"jvole.com/influx/util"
	"jvole.com/influx/weighted"
)

var (
	cfg         *config
	defaultFile = "server.yaml"
)

type config struct {
	bind        string //绑定地址端口 :8000
	buffer      uint16 //缓存
	poll        int    //轮询算法
	user        string
	password    string
	maxconn     int                    //最大连接数
	logsev      loginflux.InfluxServer //日志插件
	printscreen bool                   //日志插件是否同时输出到屏幕
	loglevel    string
}

func main() {
	file := flag.String("f", "./"+defaultFile, "config file path")
	flag.Parse()
	exist, err := util.PathExists(*file)
	if exist {
		loadConfig(*file)
	} else {
		fmt.Println("找不到配置文件，请加参数,eg：-f /etc/server.yaml")
		os.Exit(0)
	}
	serv := cfg.logsev

	lf := loginflux.NewLoginflux(serv)
	util.KitLogger = log.NewJSONLogger(log.NewSyncWriter(lf))
	// util.KitLogger = log.NewJSONLogger(os.Stdout)
	pp.Println(cfg.loglevel)
	switch cfg.loglevel {
	case "debug":
		util.KitLogger = level.NewFilter(util.KitLogger, level.AllowDebug())
	case "info":
		util.KitLogger = level.NewFilter(util.KitLogger, level.AllowInfo())
	case "error":
		util.KitLogger = level.NewFilter(util.KitLogger, level.AllowError())
	case "all":
		util.KitLogger = level.NewFilter(util.KitLogger, level.AllowAll())
	}
	util.KitLogger = log.With(util.KitLogger, "ts", log.DefaultTimestampUTC)

	fieldKeys := []string{"method"}
	var bs route.ServerRoute
	bs = route.NewServer()
	logging := log.With(util.KitLogger, "component", "influxdb")
	bs = route.NewLoggingService(logging, bs)
	bs = route.NewInstrumentingService(
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

	httpLogger := log.With(util.KitLogger, "component", "http")

	mux := http.NewServeMux()

	mux.Handle("/influx/v2/", route.MakeHandler(bs, &httpLogger))
	http.Handle("/", accessControl(mux))
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/check", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})
	// glog.Fatal(http.ListenAndServe(cfg.port, nil))

	l, err := net.Listen("tcp", cfg.bind)
	if err != nil {
		pp.Println(err)
	}

	defer l.Close()
	l = netutil.LimitListener(l, cfg.maxconn)
	fmt.Println("运行中...")
	http.Serve(l, nil)

	//TODO 平滑重启 安全退出  官方的访问限制方案对用户并不友好，可以i自己写
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

func loadConfig(file string) {
	util.Viper.SetConfigType("yaml")
	// util.Viper.SetConfigName(".cfg")
	util.Viper.AddConfigPath(".")
	// util.Viper.AddConfigPath("/Users/derek/project/go/src/jvole.com/monitor/")
	util.Viper.SetConfigFile(file)

	err := util.Viper.ReadInConfig() // 读取配置文件
	if err != nil {                  // 加载配置文件错误
		util.Log.WithFields(logrus.Fields{
			"name": "信息",
			"err":  err,
		}).Infoln("配置文件加载失败")
		return
	}

	cfg = &config{}
	mapConfig()

	//本机文件记录
	db.RecordE = db.NewBoltDB()
	//恢复进程
	recover := recover.NewWatchFail()
	// for _, v := range route.Groups {
	// 	pp.Println(len(v.S))
	// }
	recover.Run()
	//检查服务器状态
	var ig route.InfluxGateway
	ig = new(route.Group)
	checked := ig.CheckConfig()
	if !checked {
		fmt.Println("influx服务器配置或服务有问题，请检查")
		os.Exit(0)
	}
	//监控配置文件变化
	util.Viper.WatchConfig()
	util.Viper.OnConfigChange(func(e fsnotify.Event) {
		util.Log.WithFields(logrus.Fields{
			"name": "信息",
		}).Infoln("配置文件变更，重新生效")
		mapConfig()
		recover.SetServer()
		// pp.Println(cfg.buffer)
	})
}

func mapConfig() { //配置文件

	lf := util.Viper.Get("loginflux").(map[string]interface{})
	cfg.printscreen = lf["printscreen"].(bool)
	//更新influx日志插件是否显示屏幕
	loginflux.PrintScreen = cfg.printscreen

	cfg.loglevel = util.Viper.Get("loglevel").(string)

	cfg.logsev.Addr = lf["addr"].(string)
	cfg.logsev.User = lf["user"].(string)
	cfg.logsev.Passwd = lf["passwd"].(string)
	cfg.logsev.Database = lf["database"].(string)
	cfg.logsev.Table = lf["table"].(string)
	cfg.logsev.Precision = lf["precision"].(string)
	cfg.logsev.Buff = uint16(lf["buff"].(int))

	tags := util.Viper.GetStringMap("loginflux")["tags"].(map[string]interface{})
	cfg.logsev.Tags = make(map[string]string)
	cfg.logsev.Tags["server"] = tags["server"].(string)
	cfg.logsev.Tags["service"] = tags["service"].(string)

	keys := util.Viper.GetStringMap("loginflux")["keys"].([]interface{})

	for _, v := range keys {
		cfg.logsev.KeyS = append(cfg.logsev.KeyS, v.(string))
	}
	// pp.Println(cfg.logsev)
	if v, ok := util.Viper.Get("maxconn").(int); ok {
		cfg.maxconn = v
	}

	if v, ok := util.Viper.Get("user").(string); ok {
		cfg.user = v
	}
	//更新用户名
	route.User = cfg.user
	// pp.Print(route.Groups)
	if v, ok := util.Viper.Get("password").(string); ok {
		cfg.password = v
	}
	//更新密码
	route.Password = cfg.password
	if v, ok := util.Viper.Get("bind").(string); ok {
		cfg.bind = v
	}
	// pp.Print(route.Groups)
	if v, ok := util.Viper.Get("poll").(int); ok {
		cfg.poll = v
	}
	if v, ok := util.Viper.Get("buffer").(int); ok {
		cfg.buffer = uint16(v)
	}
	db.Buffer = cfg.buffer //更新缓存设置
	dbgroup := util.Viper.Get("db")
	route.Total = uint64(len(dbgroup.([]interface{})))
	route.Groups = make([]route.Group, route.Total)

	for k, v := range dbgroup.([]interface{}) { //库
		if g, ok := v.(map[interface{}]interface{}); ok { //组
			for _, gv := range g {
				// s, _ := json.Marshal(gv.([]interface{})[0])
				n := gv.([]interface{})[0].(map[interface{}]interface{})["num"]

				route.Groups[k].Num = uint64(n.(int))
				var w weighted.W
				switch cfg.poll {
				case 1:
					w = &weighted.W1{} //nginx 加权轮询算法 W2为lvs算法   W1更平滑  W2更快
				case 2:
					w = &weighted.W1{} //nginx 加权轮询算法 W2为lvs算法   W1更平滑  W2更快
				}
				route.Groups[k].Poll = w
				s := gv.([]interface{})[1].(map[interface{}]interface{})["server"].([]interface{})
				sarr := make([]route.Server, len(s))
				for sk, sv := range s { //server

					t := sv.(map[interface{}]interface{})["server"].([]interface{})
					// pp.Println(t)
					sarr[sk].Addr = t[0].(map[interface{}]interface{})["addr"].(string)
					sarr[sk].Dbname = t[3].(map[interface{}]interface{})["dbname"].(string)
					sarr[sk].Username = t[1].(map[interface{}]interface{})["username"].(string)
					sarr[sk].Password = t[2].(map[interface{}]interface{})["password"].(string)
					sarr[sk].Precision = t[4].(map[interface{}]interface{})["precision"].(string)
					sarr[sk].Weight = t[5].(map[interface{}]interface{})["weight"].(int)
					sarr[sk].QueryCount = 0
					sarr[sk].Status = true
					sarr[sk].Conn = db.NewInfluxdb(sarr[sk].Addr, sarr[sk].Username, sarr[sk].Password, sarr[sk].Dbname, sarr[sk].Precision)
					w.Add(sarr[sk].Conn, sarr[sk].Weight)
					// pp.Println(sv.(map[interface{}]interface{})["server1"])
				}
				route.Groups[k].S = sarr

			}
		}
	}

	//变更日志组件等级
	switch cfg.loglevel {
	case "debug":
		util.KitLogger = level.NewFilter(util.KitLogger, level.AllowDebug())
	case "info":
		util.KitLogger = level.NewFilter(util.KitLogger, level.AllowInfo())
	case "error":
		util.KitLogger = level.NewFilter(util.KitLogger, level.AllowError())
	case "all":
		util.KitLogger = level.NewFilter(util.KitLogger, level.AllowAll())
	}

	// util.KitLogger = log.With(util.KitLogger, "ts", log.DefaultTimestampUTC)

}
