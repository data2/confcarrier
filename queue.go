package main

import (
	"github.com/go-redis/redis"
	"log"
	"sync"
	"time"
)

func PublishMessage(rdb *redis.Client , record Record)  {
	res, err := rdb.Publish("queue:confcarrier", ToJsonString(record)).Result()
	if err != nil {
		log.Println(err,res)
		time.Sleep(time.Duration(2)*time.Second)
		PublishMessage(rdb, record)
	}
}


func SubscribeMessage(rdb *redis.Client, queue sync.Map) {
	pubSub := rdb.Subscribe("queue:confcarrier")
	ch := pubSub.Channel()
	for msg := range ch{
		recordJsonstr := msg.Payload
		BroadcastByNamespace(ToInterface(recordJsonstr).Namespace, recordJsonstr)
	}
}

