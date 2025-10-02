package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

// 创建用户的Api
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	go user.ListenMessage()
	return user
}

// 用户上线
func (this *User) Online() {
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()
	//广播当前用户上线的方法进行通知
	this.server.BroadCast(this, "已上线")
}

// 用户下线
func (this *User) Offline() {
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()
	//广播当前用户下线的方法进行通知
	this.server.BroadCast(this, "已下线")
}

// 给当前用户对应的客户端发送消息
func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg + "\n"))
}

// 用户处理消息
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线"
			this.SendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		newName := strings.Split(msg, "|")[1]
		this.server.mapLock.Lock()
		//判断新用户名是否存在
		if _, ok := this.server.OnlineMap[newName]; ok {
			this.SendMsg("该用户名已存在")
			return
		}
		// 用户名不存在，更新用户名
		delete(this.server.OnlineMap, this.Name)
		this.Name = newName
		this.server.OnlineMap[newName] = this
		this.server.mapLock.Unlock()
		this.SendMsg("您已成功重命名为" + newName)
	} else {
		this.server.BroadCast(this, msg)
	}
}

func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}
