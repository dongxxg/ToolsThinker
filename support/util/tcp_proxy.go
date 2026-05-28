package util

import (
	"flag"
	"fmt"
	"io"
	"net"
	"tools-thinker/support/logger"
)

var listenIp string
var proxyIp string

func main() {
	flag.StringVar(&listenIp, "l", "19999", "-l=0.0.0.0:9897 指定服务监听的端口")
	flag.StringVar(&proxyIp, "d", "", "-d=127.0.0.1:1789")
	flag.Parse()
	if len(proxyIp) <= 0 {
		logger.Fatal("后端ip和端口不能空,或者无效")
	}
	Server()
}

func Server() {
	lis, err := net.Listen("tcp", listenIp)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer lis.Close()
	for {
		conn, err := lis.Accept()
		if err != nil {
			fmt.Printf("建立连接错误:%v\n", err)
			continue
		}
		fmt.Println(conn.RemoteAddr(), conn.LocalAddr())
		go Handle(conn)
	}
}

func Handle(sconn net.Conn) {
	defer sconn.Close()
	dconn, err := net.Dial("tcp", proxyIp)
	if err != nil {
		fmt.Printf("连接%v失败:%v\n", proxyIp, err)
		return
	}
	exitchan := make(chan bool, 1)
	go func(sconn net.Conn, dconn net.Conn, exit chan bool) {
		_, err := io.Copy(dconn, sconn)
		fmt.Printf("往%v发送数据失败:%v\n", proxyIp, err)
		exitchan <- true
	}(sconn, dconn, exitchan)
	go func(sconn net.Conn, dconn net.Conn, exit chan bool) {
		_, err := io.Copy(sconn, dconn)
		fmt.Printf("从%v接收数据失败:%v\n", proxyIp, err)
		exitchan <- true
	}(sconn, dconn, exitchan)
	<-exitchan
	dconn.Close()
}
