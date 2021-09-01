package main

import (
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net"
	"sync"
)

const (
	GET    = "get"
	GETALL = "getall"
	SET    = "set"
	DEL    = "del"
	DELALL = "delall"
	NOTIFY = "notify"
)

type Response struct {
	Action  string
	Code    int
	Message string
	Data    interface{}
}

type Record struct {
	ID        string `gorm:"primaryKey"`
	Namespace string
	Path      string
	Value     string
}

const (
	SUCCESS = 0
	FAIL    = 1
	UNAUTH  = 2
)

func (r Response) Return(conn *net.TCPConn, action string) {
	r.Action = action
	message := ToJsonString(r)
	fmt.Println(message)
	_, err := conn.Write([]byte(message + "\n"))
	if err != nil {
		fmt.Println("send message fail ", err)
	}
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
	defer lock.Unlock()
	lock.Lock()
	response := LoadData(db, namespace, path)
	var err error
	var record Record
	if response.Code == SUCCESS && response.Data != (Record{}) {
		record = response.Data.(Record)
		err = db.Model(&(record)).Where("path", path).Update("value", value).Error
		if err != nil {
			fmt.Println("update data fail > " + err.Error())
			return Response{Code: FAIL, Message: "update data fail. "}
		}
	} else {
		uid, _ := uuid.NewUUID()
		record = Record{Namespace: namespace, ID: uid.String(), Path: path, Value: value}
		err = db.Create(&record).Error
		if err != nil {
			fmt.Println("set data fail > " + err.Error())
			return Response{Code: FAIL, Message: "set data fail. "}
		}
	}
	return Response{Code: SUCCESS, Data: record}
}

func LoadData(db *gorm.DB, namespace string, path string) Response {
	var record Record
	err := db.Where("namespace = ? and path = ?", namespace, path).Find(&record).Error
	if err != nil {
		return Response{Code: FAIL, Message: err.Error()}
	}
	return Response{Code: SUCCESS, Data: record}
}

func DelAllData(db *gorm.DB, namespace string) Response {
	err := db.Where(" namespace = ?", namespace).Delete(Record{}).Error
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
