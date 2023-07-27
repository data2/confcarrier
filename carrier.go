package main

import (
	"bufio"
	"container/list"
	"fmt"
	"github.com/go-redis/redis"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net"
	"strings"
	"sync"
	"time"
)

type ConnMeta struct {
	RemoteAddr  string
	AuthOkToken string
	Namespace   string
}

var clientMeta sync.Map
var pushQueue sync.Map
var queueMu sync.Mutex

// @description delete carrier
func RemoveCarrier(conn *net.TCPConn) {
	clientMeta.Delete(conn)
	pushQueue.Delete(conn)
	fmt.Println(fmt.Sprintf("delete client [%s] conn", conn.RemoteAddr()))
}

func HandlerClient(db *gorm.DB, conn *net.TCPConn) {
	reader := bufio.NewReader(conn)
	defer func() {
		conn.Close()
		RemoveCarrier(conn)
	}()
	for true {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(fmt.Sprintf("read req message error: %s", err))
			break
		}
		if message == "\n" {
			continue
		}
		fmt.Println(fmt.Sprintf("read req message: %s, client: %s", message, conn.RemoteAddr()))
		if len(message) == 0 {
			continue
		}
		message = strings.Replace(message, "\n", "", -1)
		info := strings.Split(message, "|")
		if len(info) < 3 {
			Response{Code: FAIL, Message: "req message not valid , must contain at least three val namespace|token|action "}.Return(conn, "")
			continue
		}

		ok := ValidCheck(conn, info[0], info[1])
		if !ok {
			Response{
				Code: UNAUTH, Message: "not valid request, pleast regist your client. ",
			}.Return(conn, info[2])
			continue
		}
		namespace := info[0]
		switch info[2] {
		case GETALL:
			LoadAllData(db, namespace).Return(conn, info[2])
		case GET:
			if len(info) != 4 {
				Response{Code: FAIL, Message: "req param valid, get must contain namespace|token|action|path"}.Return(conn, info[2])
				continue
			}
			LoadData(db, namespace, info[3]).Return(conn, info[2])
		// case SET:
		// 	if len(info) != 5 {
		// 		Response{Code: FAIL, Message: "req param valid, get must contain namespace|token|action|path|value"}.Return(conn, info[2])
		// 		continue
		// 	}
		// 	SetData(db, namespace, info[3], info[4]).Return(conn, info[2])
		// case DEL:
		// 	if len(info) != 4 {
		// 		Response{Code: FAIL, Message: "req param valid, get must contain namespace|token|action|path"}.Return(conn, info[2])
		// 		continue
		// 	}
		// 	DelData(db, namespace, info[3]).Return(conn, info[2])
		// case DELALL:
		// 	DelAllData(db, namespace).Return(conn, info[2])
		// }
	}
}

func RegistClient(conn *net.TCPConn, namespace string, token string) bool {
	bol := AuthToken(namespace, token)
	fmt.Println(bol)
	if bol {
		clientMeta.Store(conn, ConnMeta{
			RemoteAddr:  conn.RemoteAddr().String(),
			AuthOkToken: token,
			Namespace:   namespace,
		})
		Push(namespace, conn)
		return true
	}
	return false
}

func Push(namespace string, conn *net.TCPConn) {
	defer queueMu.Unlock()
	queueMu.Lock()
	val, _ := pushQueue.Load(namespace)
	if val == nil {
		subQueue := list.New()
		subQueue.PushFront(conn)
		pushQueue.Store(namespace, subQueue)
	} else {
		val.(*list.List).PushFront(conn)
	}
}

// @description auth token
func AuthToken(namespace string, token string) bool {
	if token == md5go(namespace+"666") {
		return true
	} 
	return false
}

func ValidCheck(conn *net.TCPConn, namespace string, token string) bool {
	value, _ := clientMeta.Load(conn)
	if value == nil {
		return RegistClient(conn, namespace, token)
	}
	return value.(ConnMeta).AuthOkToken == token
}

func Broadcast(s string) {
	clientMeta.Range(func(key, value interface{}) bool {
		c := key.(*net.TCPConn)
		c.Write([]byte(s))
		c.Close()
		return true
	})
}

var db *gorm.DB

func main() {
	fmt.Println("----------------------------------------------------")
	fmt.Println("confcarrier starting...")
	port := "8086"
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

	//dsn := os.Args[2]
	dsn := "ssp-app:2aGkj^Ac5c#*hZ!4@tcp(106.14.71.115:3306)/nmy-db?charset=utf8mb4&parseTime=True&loc=Local"
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
		Broadcast("exiting.")
	}()

	redisAddr := "47.100.76.173:6379"
	//redisAddr := os.Args[3]
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		Password: "4Utzy0d5p1IXeght",
	})
	pong, err := rdb.Ping().Result()
	if err != nil {
		fmt.Println(fmt.Sprintf("redis connect error : %s", err))
		return
	}
	fmt.Println("redis server connected! ping > " + pong)
	go SubscribeMessage(rdb, &pushQueue)

	for true {
		fmt.Println("tcp server accept...")
		conn, err := tcpListener.AcceptTCP()
		if err != nil {
			fmt.Println(fmt.Sprintf("accpet request from client error, %s", err))
			continue
		}
		fmt.Println(fmt.Sprintf("accpet request from client [%s] ok, join carrier ok", conn.RemoteAddr().String()))
		go HandlerClient(db, conn)
	}

	//go run carrier.go  util.go db.go queue.go 8086  "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local" "localhost:6379"
}
