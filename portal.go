package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
	"time"
)

func HttpGet(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	namespace := parseParam(r, "namespace")
	path := parseParam(r, "path")

	if len(namespace) == 0 || len(path) == 0 {
		w.Write([]byte(ToJsonString(Response{
			Code: FAIL,
			Data: "param must give",
		})))
		return
	} else if res := LoadData(httpDb, namespace, path); res.Code != SUCCESS {
		w.Write([]byte(ToJsonString(Response{
			Code: FAIL,
			Data: res.Message,
		})))
		return
	} else {
		w.Write([]byte(ToJsonString(Response{
			Code: SUCCESS,
			Data: res.Data,
		})))
	}
}

func parseParam(r *http.Request, key string) string {
	if r.Method == "GET" {
		return r.URL.Query()[key][0]
	} else {
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err)
		}
		return r.Form[key][0]
	}
}

func HttpGetAll(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	namespace := parseParam(r, "namespace")

	if len(namespace) == 0 {
		w.Write([]byte(ToJsonString(Response{
			Code: FAIL,
			Data: "param must give",
		})))
		return
	}
	if res := LoadAllData(httpDb, namespace); res.Code != SUCCESS {
		w.Write([]byte(ToJsonString(Response{
			Code: FAIL,
			Data: res.Message,
		})))
		return
	} else {
		w.Write([]byte(ToJsonString(Response{
			Code: SUCCESS,
			Data: res.Data,
		})))
	}
}

func HttpSet(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	namespace := parseParam(r, "namespace")
	path := parseParam(r, "path")
	value := parseParam(r, "value")

	if len(namespace) == 0 || len(path) == 0 || len(value) == 0 {
		w.Write([]byte(ToJsonString(Response{
			Code: FAIL,
			Data: "param must give",
		})))
		return
	}
	if res := SetData(httpDb, namespace, path, value); res.Code != SUCCESS {
		w.Write([]byte(ToJsonString(Response{
			Code: FAIL,
			Data: res.Message,
		})))
		return
	} else {
		w.Write([]byte(ToJsonString(Response{
			Code: SUCCESS,
		})))
		go PublishMessage(httpRdb, Record{Namespace: namespace, Path: path, Value: value})
	}
}

func BroadcastUpdate(namespace string, path string, value string) {
	//val, _ := pushQueue.Load(namespace)
	//for ele := val.(*list.List).Front(); ele != nil; ele = ele.Next() {
	//	Response{
	//		Code: SUCCESS, Data: Record{Namespace: namespace, Path: path, Value: value},
	//	}.Response(ele.Value.(*net.TCPConn), NOTIFY)
	//}

}

func HttpDel(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	namespace := parseParam(r, "namespace")
	path := parseParam(r, "path")

	if len(namespace) == 0 || len(path) == 0 {
		w.Write([]byte(ToJsonString(Response{
			Code: FAIL,
			Data: "param must give",
		})))
		return
	}
	if res := DelData(httpDb, namespace, path); res.Code != SUCCESS {
		w.Write([]byte(ToJsonString(Response{
			Code: FAIL,
			Data: res.Message,
		})))
		return
	} else {
		w.Write([]byte(ToJsonString(Response{
			Code: SUCCESS,
		})))
	}
}

func HttpDelAll(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	namespace := parseParam(r, "namespace")

	if len(namespace) == 0 {
		w.Write([]byte(ToJsonString(Response{
			Code: FAIL,
			Data: "param must give",
		})))
		return
	}
	if res := DelAllData(httpDb, namespace); res.Code != SUCCESS {
		w.Write([]byte(ToJsonString(Response{
			Code: FAIL,
			Data: res.Message,
		})))
		return
	} else {
		w.Write([]byte(ToJsonString(Response{
			Code: SUCCESS,
		})))
	}
}

func HttpServer(httpPort string) {
	http.HandleFunc("/get", HttpGet)
	http.HandleFunc("/getall", HttpGetAll)
	http.HandleFunc("/set", HttpSet)
	http.HandleFunc("/del", HttpDel)
	http.HandleFunc("/delall", HttpDelAll)
	err := http.ListenAndServe(":"+httpPort, nil)
	if err != nil {
		fmt.Println("httpserver start fail")
		return
	}
}

var httpDb *gorm.DB
var httpRdb *redis.Client

func main() {
	fmt.Println("----------------------------------------------------")
	fmt.Println("confcarrier-portal starting...")

	//dsn := os.Args[2]
	dsn := "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	httpDb, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,   // DSN data source name
		DefaultStringSize:         256,   // string 类型字段的默认长度
		DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false, // 根据当前 MySQL 版本自动配置
	}), &gorm.Config{})
	if err != nil {
		fmt.Println(fmt.Sprintf("mysql server connect error : %s", err))
		return
	}
	fmt.Println("mysql server connected!")

	httpDb.AutoMigrate(&Record{})
	sqlDB, err := httpDb.DB()
	if err != nil {
		fmt.Println("thread pool fail")
		return
	}
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(time.Hour)

	redisAddr := "localhost:6379"
	//redisAddr := os.Args[3]
	httpRdb = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	pong ,err:= httpRdb.Ping().Result()
	if err != nil {
		fmt.Println(fmt.Sprintf("redis connect error : %s", err))
		return
	}
	fmt.Println("redis server connected! ping > " + pong)

	httpPort := "8081"
	//httpPort := os.Args[1]

	go HttpServer(httpPort)
	fmt.Println(fmt.Sprintf("http server started, http://localhost:%s/get", httpPort))

	// go run portal.go 8081 "localhost:6379" "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
}
