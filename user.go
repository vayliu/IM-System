package main

import (
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

// NewUser 创建一个用户的 API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,

		server: server,
	}

	// 启动监听当前 User channel 消息的 goroutine
	go user.ListMessage()

	return user
}

// Online 用户上线的业务
func (u *User) Online() {
	// 1. 将用户加入到 onlineMap 中
	u.server.mapLock.Lock()
	u.server.OnlineMap[u.Name] = u
	u.server.mapLock.Unlock()
	// 2. 广播当前用户上线信息
	u.server.BroadCast(u, "已上线")
}

// Offline 用户下线的业务
func (u *User) Offline() {
	// 1. 将用户从 onlineMap 中删除
	u.server.mapLock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.mapLock.Unlock()
	// 2. 广播当前用户下线消息
	u.server.BroadCast(u, "下线")
}

// DoMessage 用户处理消息的业务
func (u *User) DoMessage(msg string) {
	if msg == "who" { // 查询当前用户都有哪些
		u.server.mapLock.Lock()
		for _, user := range u.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线...\n"
			u.SendMessage(onlineMsg)
		}
		u.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" { // 更新当前用户的用户名
		// 消息格式：rename|张三
		newName := strings.Split(msg, "|")[1]
		// 判断 newName 是否存在
		_, ok := u.server.OnlineMap[newName]
		if ok {
			u.SendMessage("当前用户名被使用\n")
		} else {
			u.server.mapLock.Lock()
			// 更新用户名
			delete(u.server.OnlineMap, u.Name)
			u.server.OnlineMap[newName] = u
			u.server.mapLock.Unlock()

			u.Name = newName
			u.SendMessage("已将您的用户名更改为：" + newName + "\n")
		}
	} else {
		u.server.BroadCast(u, msg)
	}
}

// ListMessage 监听当前 User channel 的方法，一旦有消息，就直接发送给对端客户端
func (u *User) ListMessage() {
	for {
		msg := <-u.C
		u.conn.Write([]byte(msg + "\n"))
	}
}

// SendMessage 给当前 user 对应的客户端发消息
func (u *User) SendMessage(msg string) {
	u.conn.Write([]byte(msg))
}
