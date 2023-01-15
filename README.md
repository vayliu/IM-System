# Golang 即时通信系统

## 项目介绍

本项目主要实现了一个简易的通信系统，主要包括以下功能：
 - 用户上线提醒
 - 用户消息广播
 - 查询在线用户
 - 用户私聊
 - 超时强踢

## 运行方式（以 Unix 为例）

### 服务端启动方法

1. 构建服务程序
    ```Shell
    go build -o server main.go server.go user.go
    ```
2. 启用服务程序
    ```Shell
    ./server
    ```

### 客户端启动方法

1. 构建客户端程序
    ```Shell
    go build -o client client.go
    ```
2. 启用服务程序
    ```Shell
    ./client
    ```