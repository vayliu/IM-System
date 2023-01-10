package main

import (
	"fmt"
	"net"
)

type Server struct {
	Ip   string
	Port int
}

// NewServer 创建一个 Server 接口
func NewServer(ip string, port int) *Server {
	s := &Server{
		Ip:   ip,
		Port: port,
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
}
