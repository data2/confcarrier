// @description package main
package main

import (
	"container/list"
	"github.com/go-redis/redis"
	"log"
	"net"
	"sync"
	"time"
)

func PublishMessage(rdb *redis.Client, record Record) {
	res, err := rdb.Publish("queueconfcarrier", ToJsonString(record)).Result()
	if err != nil {
		log.Println(err, res)
		time.Sleep(time.Duration(2) * time.Second)
		PublishMessage(rdb, record)
	}
}

func SubscribeMessage(rdb *redis.Client, m *sync.Map) {
	pubSub := rdb.Subscribe("queueconfcarrier")
	ch := pubSub.Channel()
	for msg := range ch {
		record := ToInterface(msg.Payload)
		val, _ := m.Load(record.Namespace)
		if val == nil {
		} else {
			queue := val.(*list.List)
			for i := queue.Front(); i != nil; i = i.Next() {
				c := i.Value.(*net.TCPConn)
				if c != nil {
					Response{
						Code: SUCCESS,
						Data: record,
					}.Return(c, NOTIFY)
				}
			}
		}
	}
}
