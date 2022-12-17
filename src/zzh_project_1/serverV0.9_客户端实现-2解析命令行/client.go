package main

import (
	"flag"
	"fmt"
	"net"
)

type Client struct {
	ServerIp   string   //连接的服务端ip
	ServerPort int      //连接的服务端port
	name       string   //当前用户名
	conn       net.Conn //当前连接的属性
}

func NewClient(serverIp string, serverPort int) *Client {
	//创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
	}

	//连接Server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}
	client.conn = conn

	//返回对象
	return client
}

var serverIp string
var serverPort int

//  执行方式： ./client -ip 127.0.0.1

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器ip地址(默认是127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口(默认是8888)")

}

func main() {
	//命令行解析
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>>连接服务器失败.....")
		return
	}
	fmt.Println(">>>>>连接服务器成功.....")
	//启动客户端业务,这里先阻塞在这，不让它死亡
	select {}
}
