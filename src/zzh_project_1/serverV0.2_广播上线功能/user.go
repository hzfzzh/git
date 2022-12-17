package main

import (
	"net"
)

type User struct {
	Name string      //用户名称
	Addr string      //用户地址ip
	C    chan string //管道channel，绑定用户的channel
	conn net.Conn    //当前用户和客户端连接的唯一标志
}

// 创建一个用户的API，方法，接口
func NewUser(conn net.Conn) *User {

	userAddr := conn.RemoteAddr().String() //从当前connection的remote连接中拿到地址，作为用户名称

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,
	}

	//启动监听当前user channel消息的goroutine
	//每次new一个user都会携带一个go程，一直监听当前user的channel
	go user.ListenMessage()

	return user

}

// 监听当前User channel的方法，一旦有消息，就直接发送给对端客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C //msg永远从user的管道里面去读，一旦有，就给它，否则就阻塞

		this.conn.Write([]byte(msg + "\n")) //
	}
}
