package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func CreateSocket() {
	tcpAdd, err := net.ResolveTCPAddr("tcp", "127.0.0.1:8086")
	if err != nil {
		fmt.Println("net.ResolveTCPAddr error:", err)
		return
	}
	conn, err := net.DialTCP("tcp", nil, tcpAdd)
	if err != nil {
		fmt.Println("net.DailTCP error:", err)
		return
	}
	defer conn.Close()
	fmt.Println("connected")
	go OnMessageRectived(conn)

	for {
		var data string
		fmt.Scan(&data)
		if data == "quit" {
			break
		}
		b := []byte(data + "\n")
		conn.Write(b)
	}
}

func OnMessageRectived(conn *net.TCPConn) {
	reader := bufio.NewReader(conn)
	for {
		// var data string
		msg, err := reader.ReadString('\n') //读取直到输入中第一次发生 ‘\n’
		fmt.Println("accept notify from server : " + msg)
		if err != nil {
			fmt.Println("err:", err)
			os.Exit(1) //服务端错误的时候，就将整个客户端关掉
		}
	}
}

func main() {
	CreateSocket()
	// namespace|b84b91beeb83bf0818bdfd9f7333f731|get|path.test
	// namespace|token|action|path
}
