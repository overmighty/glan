package client

import (
	"github.com/overmighty/glan/glanfs/api/fsapi"
	"log/slog"
	"syscall"
)

func (c *Client) do(req *fsapi.Request) *fsapi.Response {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	if err := c.conn.WriteMessage(req); err != nil {
		slog.Error("Failed to write message", "err", err, "remote_addr", c.conn.RemoteAddr())
		panic(err)
	}

	var resp fsapi.Response
	err := c.conn.ReadMessage(&resp)
	if err != nil {
		slog.Error("Failed to read message", "err", err, "remote_addr", c.conn.RemoteAddr())
		panic(err)
	}
	return &resp
}

func (c *Client) lookupNode(parentID uint64, name string) (*fsapi.LookupResponse, syscall.Errno) {
	var req fsapi.Request

	var body fsapi.LookupRequest
	body.SetParentId(parentID)
	body.SetName(name)

	req.SetLookup(&body)

	resp := c.do(&req)
	if resp.HasError() {
		return nil, fsapiErrorToErrno(resp.GetError())
	}
	// TODO: Check that resp.HasLookup().
	return resp.GetLookup(), 0
}

func (c *Client) createNode(parentID uint64, name string) (*fsapi.CreateResponse, syscall.Errno) {
	var req fsapi.Request

	var body fsapi.CreateRequest
	body.SetParentId(parentID)
	body.SetName(name)

	req.SetCreate(&body)

	resp := c.do(&req)
	if resp.HasError() {
		return nil, fsapiErrorToErrno(resp.GetError())
	}
	return resp.GetCreate(), 0
}

func (c *Client) readdirNode(id uint64) (*fsapi.ReaddirResponse, syscall.Errno) {
	var req fsapi.Request

	var body fsapi.ReaddirRequest
	body.SetId(id)

	req.SetReaddir(&body)

	resp := c.do(&req)
	if resp.HasError() {
		return nil, fsapiErrorToErrno(resp.GetError())
	}
	return resp.GetReaddir(), 0
}

func (c *Client) getattrNode(id uint64) (*fsapi.GetattrResponse, syscall.Errno) {
	var req fsapi.Request

	var body fsapi.GetattrRequest
	body.SetId(id)

	req.SetGetattr(&body)

	resp := c.do(&req)
	if resp.HasError() {
		return nil, fsapiErrorToErrno(resp.GetError())
	}
	return resp.GetGetattr(), 0
}

func (c *Client) writeFile(id uint64, data []byte, offset uint64) (*fsapi.WriteResponse, syscall.Errno) {
	var req fsapi.Request

	var body fsapi.WriteRequest
	body.SetId(id)
	body.SetData(data)
	body.SetOffset(offset)

	req.SetWrite(&body)

	resp := c.do(&req)
	if resp.HasError() {
		return nil, fsapiErrorToErrno(resp.GetError())
	}
	return resp.GetWrite(), 0
}

func (c *Client) readFile(id uint64, size uint64, offset uint64) (*fsapi.ReadResponse, syscall.Errno) {
	var req fsapi.Request

	var body fsapi.ReadRequest
	body.SetId(id)
	body.SetSize(size)
	body.SetOffset(offset)

	req.SetRead(&body)

	resp := c.do(&req)
	if resp.HasError() {
		return nil, fsapiErrorToErrno(resp.GetError())
	}
	return resp.GetRead(), 0
}
