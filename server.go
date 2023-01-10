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
	// 在线用户列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
	// 消息广播的 channel
	Message chan string
}

// NewServer 创建一个 Server 接口
func NewServer(ip string, port int) *Server {
	s := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return s
}

// Start 启动服务器的接口
func (s *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	// close listen socket
	defer listener.Close()

	// 启动监听 Message 的 goroutine
	go s.ListenMessage()

	for {
		// accept
		accept, err := listener.Accept()
		if err != nil {
			fmt.Println("listen accept err:", err)
			continue
		}

		// do handler
		go s.Handler(accept)
	}

}

func (s *Server) Handler(conn net.Conn) {
	// 当前链接的业务
	fmt.Println("链接建立成功")

	// 用户上线
	user := NewUser(conn)
	// 1. 将用户加入到 onlineMap 中
	s.mapLock.Lock()
	s.OnlineMap[user.Name] = user
	s.mapLock.Unlock()
	// 2. 广播当前用户上线信息
	s.BroadCast(user, "已上线")

	// 接收客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)

			if n == 0 {
				s.BroadCast(user, "下线")
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}
			// 提取用户的消息（去除'\n'）
			msg := string(buf[:n-1])
			// 将得到的消息进行广播
			s.BroadCast(user, msg)
		}
	}()

	// 当前 handler 阻塞
	select {}
}

// BroadCast 广播消息的方法
func (s *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	s.Message <- sendMsg
}

// ListenMessage 监听 Message 广播消息 channel 的 goroutine，一旦有消息就发送给全部的在线 User
func (s *Server) ListenMessage() {
	for {
		msg := <-s.Message
		// 将 msg 发送给全部在线的 User
		s.mapLock.Lock()
		for _, cli := range s.OnlineMap {
			cli.C <- msg
		}
		s.mapLock.Unlock()
	}
}
