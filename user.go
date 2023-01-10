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
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线..."
			u.SendMessage(onlineMsg)
		}
		u.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" { // 更新当前用户的用户名
		// 消息格式：rename|张三
		newName := strings.Split(msg, "|")[1]
		// 判断 newName 是否存在
		_, ok := u.server.OnlineMap[newName]
		if ok {
			u.SendMessage("当前用户名被使用")
		} else {
			u.server.mapLock.Lock()
			// 更新用户名
			delete(u.server.OnlineMap, u.Name)
			u.server.OnlineMap[newName] = u
			u.server.mapLock.Unlock()

			u.Name = newName
			u.SendMessage("已将您的用户名更改为：" + newName)
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		// 消息格式：to|张三|消息内容
		// 1. 获取对方用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			u.SendMessage("消息格式不正确，请使用 `to|用户名|消息内容` 格式")
			return
		}
		// 2. 根据用户名得到 User 对象
		remoteUser, ok := u.server.OnlineMap[remoteName]
		if !ok {
			u.SendMessage("该用户名不存在")
			return
		}
		// 3. 获取消息内容，通过对方的 User 对象将消息内容发送过去
		content := strings.Split(msg, "|")[2]
		if content == "" {
			u.SendMessage("无消息内容，请重发")
			return
		}
		remoteUser.SendMessage(u.Name + "对您说：" + content)
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
	u.conn.Write([]byte(msg + "\n"))
}
