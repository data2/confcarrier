package main

import (
	"bufio"
	"container/list"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type ConnMeta struct {
	RemoteAddr  string
	AuthOkToken string
	Namespace string
}

type Response struct {
	Action string
	Code    int
	Message string
	Data    interface{}
}

type Record struct {
	ID        string `gorm:"primaryKey"`
	Namespace string
	Path       string
	Value     string
}

const (
	SUCCESS = 0
	FAIL    = 1
)
const (
	GET    = "get"
	GETALL = "getall"
	SET    = "set"
	DEL    = "del"
	DELALL = "delall"
	REGIST   = "regist"
	AUTH   = "auth"
	NOTIFY = "notify"
)

var clientMeta sync.Map
var pushQueue sync.Map
var queueMu sync.Locker

func removeCarrier(conn *net.TCPConn) {
	clientMeta.Delete(conn)
	pushQueue.Delete(conn)
	fmt.Println(fmt.Sprintf("delete client [%s] conn", conn.RemoteAddr()))
}

func (r Response) Response(conn *net.TCPConn, action string) {
	r.Action = action
	message := ToJsonString(r)
	fmt.Println(message)
	conn.Write([]byte(message))
}

func ToJsonString(response Response) string {
	json1, _ := json.Marshal(response)
	return string(json1)
}

func handlerClient(db *gorm.DB, conn *net.TCPConn) {
	reader := bufio.NewReader(conn)
	defer func() {
		conn.Close()
		removeCarrier(conn)
	}()
	for true {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(fmt.Sprintf("read req message error: %s", err))
			continue
		}
		if message == "\n" {
			continue
		}
		fmt.Println(fmt.Sprintf("read req message: %s, client: %s", message, conn.RemoteAddr()))
		if len(message) == 0 {
			continue
		}
		info := strings.Split(message, "\\|")
		if len(info) < 3 {
			Response{Code: FAIL, Message: "req message not valid , must contain at least three val namespace|token|action "}.Response(conn, info[2])
			continue
		}

		notValid := NotValidCheck(conn,info[2])
		if notValid {
			Response{
				Code: FAIL,Message: "not valid request, pleast regist your client. ",
			}.Response(conn, info[2])
			continue
		}
		namespace := info[0]
		switch info[2] {
		case GETALL:
			LoadAllData(db, namespace).Response(conn, info[2])
		case GET:
			if len(info) != 4 {
				Response{Code: FAIL, Message: "req param valid, get must contain namespace|token|action|path"}.Response(conn, info[2])
				continue
			}
			LoadData(db, namespace, info[3]).Response(conn, info[2])
		case SET:
			if len(info) != 5 {
				Response{Code: FAIL, Message: "req param valid, get must contain namespace|token|action|path|value"}.Response(conn, info[2])
				continue
			}
			SetData(db, namespace, info[3], info[4]).Response(conn, info[2])
		case DEL:
			if len(info) != 4 {
				Response{Code: FAIL, Message: "req param valid, get must contain namespace|token|action|path"}.Response(conn, info[2])
				continue
			}
			DelData(db, namespace, info[3]).Response(conn, info[2])
		case DELALL:
			DelAllData(db, namespace).Response(conn, info[2])
		case REGIST:
			if len(info) != 3 {
				Response{Code: FAIL, Message: "req param valid, get must contain namespace|token|action"}.Response(conn, info[2])
				continue
			}
			RegistClient(conn, namespace, info[1]).Response(conn, info[2])
		}
	}
}

func RegistClient(conn *net.TCPConn, namespace string, token string) Response {
	bol := AuthToken(namespace, token)
	if bol {
		 clientMeta.Store(conn, ConnMeta{
		 	RemoteAddr: conn.RemoteAddr().String(),
		 	AuthOkToken: token,
		 	Namespace: namespace,
		 })
		 push(conn,namespace)
		 return Response{Code: SUCCESS,Message: "client regist success."}
	}else {
		return Response{Code: FAIL, Message: "client regist auth fail."}
	}
}

func push(conn *net.TCPConn, namespace string) {
	defer queueMu.Unlock()
	queueMu.Lock()
	val,_ := pushQueue.Load(namespace)
	if val == nil{
		subQueue := list.New()
		subQueue.PushFront(conn)
		pushQueue.Store(namespace,subQueue)
	}else {
		val.(*list.List).PushFront(conn)
	}
}

func md5go(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func AuthToken(namespace string, token string) bool {
	if token == md5go(namespace+"666") {
		return true
	} else {
		return false
	}
}

func NotValidCheck(conn *net.TCPConn, action string) bool {
	if action == REGIST{
		return false
	}
	value,_:= clientMeta.Load(conn)
	connMeta := value.(ConnMeta)
	if len(connMeta.AuthOkToken) != 0 {
		return false
	}
	return true

}

func DelData(db *gorm.DB, namespace string, path string) Response {
	err := db.Delete(&Record{}, "namespace =? and path = ?", namespace, path).Error
	if err != nil {
		return Response{Code: FAIL, Message: "del fail."}
	}
	return Response{Code: SUCCESS}
}

var lock = sync.Mutex{}

func SetData(db *gorm.DB, namespace string, path string, value string) Response {
	lock.Lock()
	record := LoadData(db, namespace, path)
	var err error
	if record.Code == SUCCESS && record.Data != (Record{}) {
		err = db.Model(&(record.Data)).Update(path, value).Error
	} else {
		uid, _ := uuid.NewUUID()
		err = db.Create(&Record{Namespace: namespace, ID: uid.String(), Path: path, Value: value}).Error
	}
	lock.Unlock()
	if err != nil {
		return Response{Code: FAIL, Message: "set data fail."}
	}
	return Response{Code: SUCCESS}
}

func LoadData(db *gorm.DB, namespace string, path string) Response {
	var record Record
	err := db.Where( "namespace = ? and path = ?", namespace,path).Find(&record).Error
	if err != nil {
		return Response{Code: FAIL, Message: err.Error()}
	}
	return Response{Code: SUCCESS, Data: record}
}

func DelAllData(db *gorm.DB, namespace string) Response {
	err := db.Where(" namespace = ?",namespace).Delete(Record{}).Error
	if err != nil {
		return Response{Code: FAIL, Message: "del all data fail."}
	}
	return Response{Code: SUCCESS}
}

func LoadAllData(db *gorm.DB, namespace string) Response {
	var records []Record
	err := db.Where("namespace=?", namespace).Find(&records).Error
	if err != nil {
		return Response{Code: FAIL, Message: "load all data fail."}
	}
	return Response{Code: SUCCESS, Data: records}
}

func CollectData(conn *net.TCPConn) ConnMeta {
	return ConnMeta{RemoteAddr: conn.RemoteAddr().String()}
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

func HttpGet(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	namespace := parseParam(r, "namespace")
	path := parseParam(r, "path")

	if len(namespace)==0 || len(path) == 0 {
		w.Write([]byte(ToJsonString(Response{
			Code: FAIL,
			Data: "param must give",
		})))
		return
	}else if  res := LoadData(db, namespace, path); res.Code != SUCCESS{
		w.Write([]byte(ToJsonString(Response{
			Code: FAIL,
			Data: res.Message,
		})))
		return
	}else{
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

	if len(namespace)==0  {
		w.Write([]byte(ToJsonString(Response{
			Code: FAIL,
			Data: "param must give",
		})))
		return
	}
	if  res := LoadAllData(db, namespace); res.Code != SUCCESS{
		w.Write([]byte(ToJsonString(Response{
			Code: FAIL,
			Data: res.Message,
		})))
		return
	}else{
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

	if len(namespace)==0 || len(path)==0 || len(value)==0  {
		w.Write([]byte(ToJsonString(Response{
			Code: FAIL,
			Data: "param must give",
		})))
		return
	}
	if  res := SetData(db, namespace, path,value); res.Code != SUCCESS{
		w.Write([]byte(ToJsonString(Response{
			Code: FAIL,
			Data: res.Message,
		})))
		return
	}else{
		w.Write([]byte(ToJsonString(Response{
			Code: SUCCESS,
		})))
		go broadcastUpdate(namespace, path, value)
	}
}

func broadcastUpdate(namespace string, path string, value string) {
	val, _ := pushQueue.Load(namespace)
	for ele := val.(*list.List).Front(); ele != nil; ele = ele.Next() {
		Response{
			Code: SUCCESS, Data: Record{Namespace: namespace, Path: path, Value: value},
		}.Response(ele.Value.(*net.TCPConn), NOTIFY)
	}

}

func HttpDel(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	namespace := parseParam(r, "namespace")
	path := parseParam(r, "path")

	if len(namespace)==0 || len(path)==0  {
		w.Write([]byte(ToJsonString(Response{
			Code: FAIL,
			Data: "param must give",
		})))
		return
	}
	if  res := DelData(db, namespace, path); res.Code != SUCCESS{
		w.Write([]byte(ToJsonString(Response{
			Code: FAIL,
			Data: res.Message,
		})))
		return
	}else{
		w.Write([]byte(ToJsonString(Response{
			Code: SUCCESS,
		})))
	}
}

func HttpDelAll(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	namespace := parseParam(r, "namespace")

	if len(namespace)==0  {
		w.Write([]byte(ToJsonString(Response{
			Code: FAIL,
			Data: "param must give",
		})))
		return
	}
	if  res := DelAllData(db, namespace); res.Code != SUCCESS{
		w.Write([]byte(ToJsonString(Response{
			Code: FAIL,
			Data: res.Message,
		})))
		return
	}else{
		w.Write([]byte(ToJsonString(Response{
			Code: SUCCESS,
		})))
	}
}

func broadcastExiting() {
	clientMeta.Range(func(key, value interface{}) bool {
		c := key.(*net.TCPConn)
		c.Write([]byte("config carrier exiting..."))
		c.Close()
		return true
	})
}

var db *gorm.DB

func main() {
	fmt.Println("----------------------------------------------------")
	fmt.Println("config-carrier starting...")
	port := "8081"
	//port := os.Args[1]
	fmt.Println("output param loading")
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("127.0.0.1:%s", port))
	if err != nil {
		fmt.Println("resolve ip port fail")
		return
	}
	tcpListener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		fmt.Println(fmt.Sprintf("server start fail! %s", err))
		return
	}
	fmt.Println("tcp server started!")
	//dsn := os.Args[3]
	dsn := "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,   // DSN data source name
		DefaultStringSize:         256,   // string 类型字段的默认长度
		DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false, // 根据当前 MySQL 版本自动配置
	}), &gorm.Config{})
	if err != nil {
		fmt.Println(fmt.Sprintf("db start error : %s", err))
		return
	}
	fmt.Println("mysql server connected!")

	db.AutoMigrate(&Record{})
	sqlDB, err := db.DB()
	if err != nil {
		fmt.Println("thread pool fail")
		return
	}
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(time.Hour)
	defer func() {
		tcpListener.Close()
		broadcastExiting()
	}()

	httpPort := "8081"
	//httpPort := os.Args[2]

	go HttpServer(httpPort)
	fmt.Println(fmt.Sprintf("http server started, http://localhost:%s/get", httpPort))

	for true {
		fmt.Println("tcp server accept...")
		conn, err := tcpListener.AcceptTCP()
		if err != nil {
			fmt.Println(fmt.Sprintf("accpet request from client error, %s", err))
			continue
		}
		fmt.Println(fmt.Sprintf("accpet request from client [%s] ok, join carrier ok", conn.RemoteAddr().String()))
		go handlerClient(db, conn)
	}

}
