package common

import (
	"bufio"
	"google.golang.org/protobuf/encoding/protodelim"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/runtime/protoiface"
	"net"
)

type Message interface {
	proto.Message
	protoiface.MessageV1
}

type Conn[TOut Message, TIn Message] struct {
	conn net.Conn
	bufr *bufio.Reader
}

func NewConn[TOut Message, TIn Message](conn net.Conn) *Conn[TOut, TIn] {
	return &Conn[TOut, TIn]{
		conn: conn,
		bufr: bufio.NewReader(conn),
	}
}

func (c *Conn[TOut, TIn]) WriteMessage(m TOut) error {
	//slog.Debug("WriteMessage", "remote_addr", c.RemoteAddr(), "m", m)
	_, err := protodelim.MarshalTo(c.conn, m)
	return err
}

func (c *Conn[TOut, TIn]) ReadMessage(m TIn) error {
	//defer slog.Debug("ReadMessage", "remote_addr", c.RemoteAddr(), "m", m)
	return protodelim.UnmarshalFrom(c.bufr, m)
}

func (c *Conn[TOut, TIn]) Close() error {
	return c.conn.Close()
}

func (c *Conn[TOut, TIn]) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}
