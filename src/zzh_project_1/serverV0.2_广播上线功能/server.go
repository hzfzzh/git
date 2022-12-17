package main

import (
	"fmt"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int
	//增加一个user的map表和相应的channel，这是个广播的管道
	//在线用户的列表
	OnlineMap map[string]*User //是个map，key是用户名，value是用户对象
	mapLock   sync.RWMutex     //map是一个全局的，需要加一个锁，sync包提供

	//消息广播的channel
	Message chan string
}

// 创建一个server的接口/对象
func NewServer(ip string, port int) *Server { //返回server对象，是一个指针
	server := &Server{ //new一个server对象，即把函数当前的ip和port给这个new的server
		Ip:        ip, //它本身就是一个指针，
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// 监听Message广播消息channel的goroutine，一旦有消息就发送给全部的在线User
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message //不断从这个channel读数据

		//将msg发送给全部的在线user
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap { //不关心key，只关心value，是咱们的用户*User
			cli.C <- msg //一旦当前用户的C接收到msg，那么user.go里的ListenMessage方法就可以得到消息，并把消息写给对应的客户端
		}
		this.mapLock.Unlock()

	}
}

// 写一个广播方法
// 形参是用户和消息，指是哪个用户发起的，是什么消息
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	this.Message <- sendMsg //最终把msg发送到这个this的Message中，广播消息
}

func (this *Server) Handler(conn net.Conn) { //形参即是当前创建成功的链接的套接字
	// ...当前链接的业务//这里写业务
	// fmt.Println("链接建立成功")

	//用户上线，将消息广播,将用户加入到onlineMap中
	user := NewUser(conn) //先把user搞过来，再在这里去用它

	this.mapLock.Lock() //先上锁
	this.OnlineMap[user.Name] = user
	this.mapLock.Unlock() //后解锁，这是个常见的加解锁

	//广播当前用户上线消息
	this.BroadCast(user, "one user is onlineing已上线")

	//为保证广播完后不清掉，需要阻塞在这
	//当前handler阻塞
	select {} //这里这样就可以了？

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

	//启动监听Message的goroutine
	go this.ListenMessager() //要求当server启动时就该启动监听，因此需要协程永远启动这个过程

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
