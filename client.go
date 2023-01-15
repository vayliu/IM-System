package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIP   string
	ServerPort int
	Name       string
	conn       net.Conn
	mode       int // 当前 Client 的模式
}

// NewClient 构造函数
func NewClient(serverIP string, serverPort int) *Client {
	// 创建客户端对象
	client := &Client{
		ServerIP:   serverIP,
		ServerPort: serverPort,
		mode:       999,
	}
	// 连接 Server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIP, serverPort))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}

	client.conn = conn

	// 返回对象
	return client
}

// menu 菜单显示
func (c *Client) menu() bool {
	var mode int
	fmt.Println("1. 公聊模式")
	fmt.Println("2. 私聊模式")
	fmt.Println("3. 更新用户名")
	fmt.Println("0. 退出")

	fmt.Scanln(&mode)

	if mode >= 0 && mode <= 3 {
		c.mode = mode
		return true
	} else {
		fmt.Println(">>>> 请输入合法范围内的数字 <<<<")
		return false
	}
}

// DealResponse 处理 Server 回应的消息，直接显示到标准输出
func (c *Client) DealResponse() {
	// 一旦 c.conn 有数据，就直接 copy 到 stdout 标准输出上，永久阻塞监听
	io.Copy(os.Stdout, c.conn)
}

// PublicChat 公聊模式
func (c *Client) PublicChat() {

	var chatMsg string

	// 提示用户输入消息
	fmt.Println(">>>>> 请输入聊天内容（exit 退出）：")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		// 消息不为空则发送给服务器
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := c.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write error:", err)
				break
			}
		}

		chatMsg = ""
		fmt.Println(">>>>> 请输入聊天内容（exit 退出）：")
		fmt.Scanln(&chatMsg)
	}
}

// SelectUsers 查询当前在线用户
func (c *Client) SelectUsers() {
	sendMsg := "who\n"
	_, err := c.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn Write error:", err)
		return
	}
}

// PrivateChat 私聊模式
func (c *Client) PrivateChat() {

	var remoteName string
	var chatMsg string

	// 查询当前用户
	c.SelectUsers()
	fmt.Println(">>>>> 请输入聊天对象[用户名]（exit 退出）：")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>>>> 请输入消息内容（exit 退出）：")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			// 消息不为空则发送给服务器
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n"
				_, err := c.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write error:", err)
					break
				}
			}

			chatMsg = ""
			fmt.Println(">>>>> 请输入消息内容（exit 退出）：")
			fmt.Scanln(&chatMsg)
		}

		c.SelectUsers()
		fmt.Println(">>>>> 请输入聊天对象[用户名]（exit 退出）：")
		fmt.Scanln(&remoteName)

	}
}

// UpdateName 更新用户名
func (c *Client) UpdateName() bool {

	fmt.Println(">>>>> 请输入用户名（exit 退出）：")
	fmt.Scanln(&c.Name)

	for c.Name != "exit" {
		sendMsg := "rename|" + c.Name + "\n"
		_, err := c.conn.Write([]byte(sendMsg))
		if err != nil {
			fmt.Println("conn Write error:", err)
			return false
		}
		return true
	}
	return false
}

// Run 执行
func (c *Client) Run() {
	for c.mode != 0 {
		for c.menu() != true {
		}
		// 根据不同的模式处理不同的业务
		switch c.mode {
		case 1:
			// 公聊模式
			fmt.Println("公聊模式选择...")
			c.PublicChat()
		case 2:
			// 私聊模式
			fmt.Println("私聊模式选择...")
			c.PrivateChat()
		case 3:
			// 更新用户名
			fmt.Println("更新用户名选择...")
			c.UpdateName()
		}
	}
}

var serverIP string
var serverPort int

func init() {
	flag.StringVar(&serverIP, "ip", "127.0.0.1", "设置连接服务器的 IP 地址（默认是127.0.0.1）")
	flag.IntVar(&serverPort, "port", 8888, "设置连接服务器的端口号（默认是8888）")
}

func main() {
	// 命令行解析
	flag.Parse()

	client := NewClient(serverIP, serverPort)
	if client == nil {
		fmt.Println(">>>>>>> 连接服务器失败...")
		return
	}
	fmt.Println(">>>>>>> 连接服务器成功...")

	// 单独开启一个 goroutine 去处理 Server 回执的消息
	go client.DealResponse()

	// 启动客户端的业务
	client.Run()
}
