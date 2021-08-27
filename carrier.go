package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type ConnMeta struct {
	RemoteAddr  string
	AuthOkToken string
}

type CarrierConfig struct {
	Message   map[string]string
	Namespace string
}

type Response struct {
	Code    int
	Message string
	Data    interface{}
}

type Record struct {
	ID        string `gorm:"primaryKey"`
	Namespace string
	Path      string
	Value     interface{}
	CreateAt  time.Time
	UpdatedAt time.Time
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
	AUTH   = "auth"
)

var meta map[*net.TCPConn]ConnMeta

func removeCarrier(conn *net.TCPConn) {
	delete(meta, conn)
	fmt.Println(fmt.Sprintf("delete client [%s] conn", conn.RemoteAddr()))
}

func (r Response) Response(conn *net.TCPConn) {
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
			Response{Code: FAIL, Message: "req message not valid , must contain at least three val namespace|token|action "}.Response(conn)
			continue
		}

		connMeta, needAuth := CheckNeedAuth(conn)
		if needAuth {
			res := AuthToken(connMeta, info[0], info[1])
			if !res {
				Response{Code: FAIL, Message: "auth fail, please try again."}.Response(conn)
				continue
			}
		}
		namespace := info[0]
		switch info[2] {
		case GETALL:
			LoadAllData(db, namespace).Response(conn)
		case DELALL:
			DelAllData(db, namespace).Response(conn)
		case GET:
			if len(info) != 4 {
				Response{Code: FAIL, Message: "req param valid, get must contain namespace|token|action|key"}.Response(conn)
				continue
			}
			LoadData(db, namespace, info[3]).Response(conn)
		case SET:
			if len(info) != 5 {
				Response{Code: FAIL, Message: "req param valid, get must contain namespace|token|action|key|value"}.Response(conn)
				continue
			}
			SetData(db, namespace, info[3], info[4]).Response(conn)
		case DEL:
			if len(info) != 4 {
				Response{Code: FAIL, Message: "req param valid, get must contain namespace|token|action|key"}.Response(conn)
				continue
			}
			DelData(db, namespace, info[3]).Response(conn)
		}

	}
}

func md5go(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func AuthToken(connMeta ConnMeta, namespace string, token string) bool {
	if token == md5go(namespace+"666") {
		connMeta.AuthOkToken = token
		return true
	} else {
		return false
	}
}

func CheckNeedAuth(conn *net.TCPConn) (ConnMeta, bool) {
	connMeta := meta[conn]
	if connMeta.AuthOkToken == "" {
		return connMeta, true
	}
	return connMeta, false
}

func DelData(db *gorm.DB, namespace string, key string) Response {
	db.Delete(&Record{}, "path = ?", key)
	return Response{SUCCESS, "", nil}
}

func SetData(db *gorm.DB, namespace string, key string, value string) Response {
	lock := sync.Mutex{}
	lock.Lock()
	record := LoadData(db, namespace, key)
	if record.Data != nil {
		db.Model(&(record.Data)).Update(key, value)
	} else {
		db.Create(&Record{Namespace: namespace, Path: key, Value: value})
	}
	lock.Unlock()
	return Response{SUCCESS, "", nil}
}

func LoadData(db *gorm.DB, namespace string, key string) Response {
	record := Record{}
	db.First(&record, "path = ?", key)
	return Response{Code: SUCCESS, Data: record}
}

func DelAllData(db *gorm.DB, namespace string) Response {
	db.Where("1=1").Delete(Record{})
	return Response{SUCCESS, "", nil}
}

func LoadAllData(db *gorm.DB, namespace string) Response {
	var records []Record
	db.Where("namespace=?", namespace).Find(&records)
	return Response{Code: SUCCESS, Data: records}
}

func collectData(conn *net.TCPConn) ConnMeta {
	return ConnMeta{RemoteAddr: conn.RemoteAddr().String()}
}

func broadcastExiting() {
	for tcpConn := range meta {
		tcpConn.Write([]byte("config carrier exiting..."))
		tcpConn.Close()
	}
}

func main() {
	fmt.Println("----------------------------------------------------")
	fmt.Println("config-carrier starting...")
	port := os.Args[1]
	fmt.Println("output param loading...")
	meta = make(map[*net.TCPConn]ConnMeta)
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
	// "gorm:gorm@tcp(127.0.0.1:3306)/gorm?charset=utf8&parseTime=True&loc=Local"
	dsn := os.Args[2]
	db, err := gorm.Open(mysql.New(mysql.Config{
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

	for true {
		conn, err := tcpListener.AcceptTCP()
		if err != nil {
			fmt.Println(fmt.Sprintf("accpet request from client error, %s", err))
			continue
		}
		meta[conn] = collectData(conn)
		fmt.Println(fmt.Sprintf("accpet request from client [%s] ok, join carrier ok", conn.RemoteAddr().String()))
		go handlerClient(db, conn)
	}

}
