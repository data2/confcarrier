package main

import (
	"fmt"
)

func add(x int, y int) int {
	return x + y
}

func main() {
	//	//strs := strings.Split("test|1534c9b83ac6ede0a791503ade12d72a|getall", "|")
	//	//fmt.Println(len(strs))
	//	//fmt.Println(strs[2])
	//
	//	var r Record
	//	s := `{"ID":"101a7cb6-0aff-11ec-b26b-88e9fe840b9a","Namespace":"test","Path":"kkk","Value":"uuuu1"}`
	//	json.Unmarshal([]byte(s), &r)
	//	fmt.Println(r)
	var x int = add(2, 3)
	fmt.Println(x)

	fmt.Println(md5go("namespace" + "666"))
}
