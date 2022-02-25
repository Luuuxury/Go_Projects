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
	// 在线的用户列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex //RWMutex（读写锁）

	//消息广播的channel
	Message chan string
}

// NewServer NewsServer 创建一个 server 对象，返回用指针接收
func NewServer(ip string, port int) *Server {
	// new 一个server 对象，把&Server地址指针给了server 然后让*Server 接收
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// Start 启动服务 相当于 Server类的一个成员方法
func (this *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	// close listener socket
	defer listener.Close()
	go this.ListenServerMessage()
	// accept 如果接收到连接，就去处理业务 --> this.Handler(conn)
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept:", err)
			continue
		}
		go this.Handler(conn)
	}
}

// ListenServerMessage 持续监听
func (this *Server) ListenServerMessage() {
	for {
		msg := <-this.Message
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

// Handler 处理业务
func (this *Server) Handler(conn net.Conn) {
	// 用户上线，将用户加入到OnlineMap
	user := NewUser(conn)
	this.mapLock.Lock()
	this.OnlineMap[user.Name] = user
	this.mapLock.Unlock()
	// 一个用户上线了就向其他用户广播一下
	this.BroadCast(user, "用户上线了")

	// 实现接收一个用户的消息，然后广播这个用户的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				this.BroadCast(user, "下线")
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}
			// 提取用户发过来的消息（去除'\n'）
			msg := string(buf[:n-1])
			this.BroadCast(user, msg)
		}

	}()

	// 发送完消息，先让当前 Handler 阻塞，否则他的子goroutine 都得GG
	select {}
}

// BroadCast 广播方法
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}
