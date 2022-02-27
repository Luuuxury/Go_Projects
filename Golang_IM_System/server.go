package main

import (
	"fmt"
	"io"
	"net"
	"runtime"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int
	// 在线的用户列表
	OnlineMap map[string]*User // key 值就是用户名，value 是用户对象
	mapLock   sync.RWMutex     // 因为OnlineMap 是个全局的可能需要加一个锁

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
	go this.ListenMessage()
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

// ListenMessage Server启动后持续监听 写入 clis.C 管道
func (this *Server) ListenMessage() {
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
	user := NewUser(conn, this)
	user.Online()
	isAlive := make(chan bool)
	// 实现接收一个用户的消息，然后广播这个用户的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Connect Read err", err)
				return
			}
			// 提取用户发过来的消息（去除'\n'）
			msg := string(buf[:n-1])
			user.DoMessage(msg)
			isAlive <- true
		}
	}()
	// 判断用户是否活跃
	for {
		// 只要所有 case 不执行，select 就一直阻塞，当isAlive channle 中有True 时候，第一个case执行 打破原有的 select 阻塞，重新刷新select
		select {
		case <-isAlive:
			// 什么也不用做
		case <-time.After(time.Second * 600):
			user.SendMsg("你被踢了")
			close(user.C) //销毁使用的资源
			conn.Close()
			runtime.Goexit() // return
		}
	}
}

// BroadCast 广播方法
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}
