package main

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

//func main() {
//	//list := list.List{}
//	//list.PushFront(11)
//	//go printInfo1(&list)
//	//
//	//time.Sleep(time.Duration(3)*time.Second)
//	//list.PushFront(222)
//	//
//	//time.Sleep(time.Duration(1000)*time.Second)
//
//	list := list.List{}
//	list.PushFront(22)
//	list.PushFront(2233)
//
//	for e := list.Front(); e != nil; e = e.Next() {
//		fmt.Println(e.Value)
//	}
//}

func printInfo1(list *list.List) {
	mu := sync.Mutex{}
	for true {
		mu.Lock()
		fmt.Println(list.Len())
		ele := list.Front()
		if ele != nil {
			fmt.Println(ele.Value)
			list.Remove(ele)
			fmt.Println(list.Len())
		}

		time.Sleep(time.Duration(6) * time.Second)
		mu.Unlock()

	}

}
