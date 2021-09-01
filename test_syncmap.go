package main

type Product struct {
	Name string
}
type Price struct {
	Money int
}

//
//func main() {
//	var m sync.Map
//	var p = Product{
//		Name: "111",
//	}
//	var price = Price{
//		Money: 200,
//	}
//	m.Store(p, price)
//
//	val, _ := m.Load(p)
//
//	fmt.Println(val.(Price).Money)
//
//}
