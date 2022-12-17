package main

import (
	"fmt"
	"io"
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
	user := NewUser(conn, this) //先把user搞过来，再在这里去用它//zzh555由于user.go那边改动了，所以这边要多传一个当前用户的server，这里即this

	user.Online() //zzh 11 这里就是用户的上线业务

	//接受客户端发送的消息
	//对socket进行基本类的操作，需要匿名函数去做
	//这个gofunc怎么读到用户端的消息还不是很理解
	//先创空的buf，然后用conn.Read()把结果给到buf？
	//可是这里的conn.Read(buf)，buf不是传入参数吗？
	//额，应该和conn的write和read相对应，这个需要理解下。
	//明白了，这里的buf就是Write()里的msg，它是从指针*User.C的channel里得到的，由用户输入，给到客户端，存在user结构体的指针。通过conn的write和read方法实现写读的管道传输，这里有一个读写分离的思想。
	go func() {
		buf := make([]byte, 4096) //这个buf是最大4k的切片字节
		for {
			n, err := conn.Read(buf) //read方法从当前套接字中读数据，返回字节数n//读取当前用户输入信息的字段个数，返回一个int
			if n == 0 {              //n为0代表客户端合法关闭close
				user.Offline() //zzh 12 这里就是用户下线业务
				return
			}
			if err != nil && err != io.EOF { //读完都有EOF表示文件末尾，若无代表非法
				fmt.Println("Conn Read err:", err)
				return
			} //到此为止，算能得到正常的buf
			//提取用户的消息（去除'\n')
			msg := string(buf[:n-1]) //把字节转为字符串

			//用户针对msg进行消息处理
			user.DoMessage(msg) //zzh 13 用户消息处理业务

		}
	}() //每个客户端都有这个匿名go程处理客户端 读 的业务
	//同理要求user.go里需要一个go程想客户端 写

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
