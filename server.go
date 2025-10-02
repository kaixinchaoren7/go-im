package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

// 创建服务器类
type Server struct {
	Ip   string
	Port int

	OnlineMap map[string]*User //服务器的用户列表
	mapLock   sync.RWMutex     //全局锁
	Message   chan string      // 服务器的管道

}

// 提供一个获得Server的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// 服务器channel一旦有消息就广播发送该消息
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message
		//一旦有消息发送给全部的在线用户
		this.mapLock.Lock()
		for _, value := range this.OnlineMap {
			value.C <- msg
		}
		this.mapLock.Unlock()
	}
}

// 广播发送用户上线
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg

}

// 处理当前连接的业务handler
func (this *Server) Handler(conn net.Conn) {
	fmt.Println("连接建立成功！！！")

	//将新增加的用户添加到全局用户列表中
	user := NewUser(conn, this)

	//调用用户上线的接口
	user.Online()

	//接受客户端消息 进行广播
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				//调用用户下线
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("conn.Read err is", err)
				return
			}
			// 提取用户的消息，去掉换行符
			msg := string(buf[:n-1])
			//调用用户处理消息的接口
			user.DoMessage(msg)
		}
	}()
	//当前handle阻塞
	select {}
}

// 给Server类添加一个启动的方法（该方法为成员方法）
func (this *Server) Start() {
	// 1 : 进行socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listener err is", err)
		return
	}
	defer listener.Close()

	//启动监听消息的goroutine
	go this.ListenMessager()

	for {
		// accept 进行连接
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept error:", err)
		}
		//do hanlder
		go this.Handler(conn)
	}
}
