package main

import (
	"net"
)

type User struct {
	Name string      //用户名称
	Addr string      //用户地址ip
	C    chan string //管道channel，绑定用户的channel
	conn net.Conn    //当前用户和客户端连接的唯一标志

	server *Server //zzh222当前用户属于哪个server，是为了当前用户类能访问到Server里的属性
}

// 创建一个用户的API，方法，接口
func NewUser(conn net.Conn, server *Server) *User { //zzh333这里就需要新增形参*Server把当前Server属性给传进来

	userAddr := conn.RemoteAddr().String() //从当前connection的remote连接中拿到地址，作为用户名称

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,

		server: server, //zzh444当前user的server就等于你传进来的server
	}

	//启动监听当前user channel消息的goroutine
	//每次new一个user都会携带一个go程，一直监听当前user的channel
	go user.ListenMessage()

	return user
}

// 用户的上线业务
// zzh111这里从server.go把用户上线的代码复制过来之后，明显这个this用户类是无法访问到server的map的，因此咱们需要给user结构体新增一个其所属server的属性
func (this *User) Online() {
	//用户上线，将消息广播,将用户加入到onlineMap中
	this.server.mapLock.Lock()              //先上锁	//zzh666做完前5步后，我们就能在当前user中访问当前user的map属性了
	this.server.OnlineMap[this.Name] = this //zzh6.5 user改成this，换个新名字
	this.server.mapLock.Unlock()            //后解锁，这是个常见的加解锁

	//广播当前用户上线消息
	this.server.BroadCast(this, "one user is onlineing已上线") //zzh777,把user改成this，因为上面的加解锁操作里把用户名赋值给this了
}

// 用户的下线业务
func (this *User) Offline() {
	//zzh888,把上面的上线代码直接全部拷贝到这里并做修改。
	//用户下线，将消息广播,将用户从onlineMap中删除
	this.server.mapLock.Lock() //先上锁	//zzh666做完前5步后，我们就能在当前user中访问当前user的map属性了
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock() //后解锁，这是个常见的加解锁

	//广播当前用户上线消息
	this.server.BroadCast(this, "one user is offline已下线") //zzh777,把user改成this，因为上面的加解锁操作里把用户名赋值给this了

}

// 用户处理消息的业务
func (this *User) DoMessage(msg string) {
	//zzh999 目前用户处理的业务就广播一条
	//zzh101010做完这些之后，就可以在server.go中删除原来的代码了
	this.server.BroadCast(this, msg)
}

// 监听当前User channel的方法，一旦有消息，就直接发送给对端客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C //msg永远从user的管道里面去读，一旦有，就给它，否则就阻塞

		this.conn.Write([]byte(msg + "\n")) //这里和server.go里面的conn.Read(buf)功能相对应
	}
}
