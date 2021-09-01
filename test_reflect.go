package main

import (
	"reflect"
)

func dynamic(f interface{}, args []interface{}) []reflect.Value {
	fun := reflect.ValueOf(test)
	param := make([]reflect.Value, len(args))
	for k, p := range args {
		param[k] = reflect.ValueOf(p)
	}
	return fun.Call(param)
}

//func main()  {
//	p := []interface{}{2}
//	res := dynamic("test",p)
//	for _,v := range res{
//		fmt.Println(v.Int())
//	}
//}

func test(a int) int {
	return a + 1
}
