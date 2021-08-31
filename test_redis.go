package main

import (
	"fmt"
	"github.com/go-redis/redis"
)

func main()  {
	rdb := redis.NewClient(&redis.Options{Addr: ""})
	pong, _ := rdb.Ping().Result()
	fmt.Println(pong)
	rdb.Close()
}
