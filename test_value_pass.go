package main

import (
	"sync"
	"time"
)

//func main()  {
//	m := sync.Map{}
//	m.Store("1","2")
//	go printInfo(m)
//
//	time.Sleep(time.Duration(3)*time.Second)
//	m.Store("2","3")
//}

func printInfo(m sync.Map) {
	for true {
		m.Range(func(key, value interface{}) bool {
			println(key)
			return true
		})
		time.Sleep(time.Duration(3) * time.Second)

	}

}
