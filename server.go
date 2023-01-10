package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
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

	// 将当前连接信息并封装成用户
	user := NewUser(conn, s)
	// 用户上线
	user.Online()

	// 监听用户是否活跃的 channel
	isLive := make(chan bool)

	// 接收客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)

			if n == 0 {
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			// 提取用户的消息（去除'\n'）
			msg := string(buf[:n-1])

			// 用户针对 msg 进行消息处理
			user.DoMessage(msg)

			// 用户的任意消息代表当前用户是活跃的
			isLive <- true
		}
	}()

	// 当前 handler 阻塞
	for {
		select {
		case <-isLive:
			// 当前用户是活跃的，应重置定时器
			// 不做任何操作，为了激活 select，更新下面的定时器
		case <-time.After(5 * 60 * time.Second): // 已经超时 10 秒
			// 将当前的 User 强制关闭
			user.SendMessage("你被踢了")
			// 销毁用户资源
			close(user.C)
			// 关闭连接
			conn.Close()
			// 退出当前的 handler
			return // runtime.Goexit()
		}
	}
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
