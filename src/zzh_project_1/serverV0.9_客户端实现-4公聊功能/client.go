package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string   //连接的服务端ip
	ServerPort int      //连接的服务端port
	Name       string   //当前用户名
	conn       net.Conn //当前连接的属性
	flag       int      //当前client的模式
}

func NewClient(serverIp string, serverPort int) *Client {
	//创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999, //设置默认值为999因为一旦为0后面设置直接退出
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

// 处理server端回应的消息的go程，直到显示到标准输出即可
func (client *Client) DealResponse() {
	//一旦client.conn有数据，就直接copy到stdout标准输出上，永久阻塞监听
	io.Copy(os.Stdout, client.conn)
	//这句话等价于下面的屏蔽代码
	// for {
	// 	buf := make()
	// 	client.conn.Read(buf) //读到buf中
	// 	fmt.Println("buf",buf)  //直接把buf打到屏幕上
	// }
}

func (client *Client) menu() bool { //绑定一个菜单的方法
	var flag int
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true //合法返回true，非法返回false
	} else {
		fmt.Println(">>>>>请输入合法范围内的数字<<<<<")
		return false
	}
}

func (client *Client) PublicChat() {
	//提示用户输入信息
	var chatMsg string
	fmt.Println(">>>>>请输入聊天内容,exit退出")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		//发送给服务器
		//消息不为空则发送
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write err:", err)
				break
			}
		}
		chatMsg = ""
		fmt.Println(">>>>>请输入聊天内容,exit退出")
		fmt.Scanln(&chatMsg)
	}
}

func (client *Client) UpdateName() bool {
	fmt.Println(">>>>>请输入用户名:")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}
	return true
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true { //不断提示用户菜单，直到用户输入使值为true
		}
		//根据不同的模式处理不同的业务
		switch client.flag {
		case 1:
			//公聊模式
			client.PublicChat()
			break
		case 2:
			//私聊模式
			fmt.Println("私聊模式选择...")
			break
		case 3:
			//更新用户名
			client.UpdateName()
			break
		}
	}
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

	//单独开启一个goroutine，去处理server端的回执消息
	go client.DealResponse()

	fmt.Println(">>>>>连接服务器成功.....")
	//启动客户端业务,这里先阻塞在这，不让它死亡
	client.Run()
}
