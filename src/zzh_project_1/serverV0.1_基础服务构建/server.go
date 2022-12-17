package main

import (
	"fmt"
	"net"
)

type Server struct {
	Ip   string
	Port int
}

// 创建一个server的接口/对象
func NewServer(ip string, port int) *Server { //返回server对象，是一个指针
	server := &Server{ //new一个server对象，即把函数当前的ip和port给这个new的server
		Ip:   ip, //它本身就是一个指针，
		Port: port,
	}
	return server
}

func (this *Server) Handler(conn net.Conn) { //形参即是当前创建成功的链接的套接字
	// ...当前链接的业务//这里写业务
	fmt.Println("链接建立成功")
}

// 启动服务器的接口
func (this *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}

	//close listen socket
	defer listener.Close() //为遗忘close，先defer掉

	for {
		//accept		//明显accept和handle都应该在循环中，因为服务器它一直要监听和处理
		conn, err := listener.Accept() //accept成功，说明当前有一个客户端已经过来了
		if err != nil {                //建立成功之后，这个conn就是当前客户端的一个套接字
			fmt.Println("listener accept err:", err)
			continue
		}
		//do handler//go的一个业务
		//为保证不阻塞，需要协程启动，不耽误下一次accept
		go this.Handler(conn) //开辟协程后，主go程会循环下一个accept
	}

}
