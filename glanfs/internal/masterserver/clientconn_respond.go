package masterserver

import (
	"github.com/overmighty/glan/glanfs/api/fsapi"
	"log/slog"
)

func (c *clientConn) respond(resp *fsapi.Response) {
	if err := c.conn.WriteMessage(resp); err != nil {
		slog.Error("Failed to write message", "err", err, "remote_addr", c.conn.RemoteAddr())
		panic(err)
	}
}

func (c *clientConn) respondError(err fsapi.Error) {
	var resp fsapi.Response

	resp.SetError(err)

	c.respond(&resp)
}

func (c *clientConn) respondLookup(id uint64, typ fsapi.FileType) {
	var resp fsapi.Response

	var body fsapi.LookupResponse
	body.SetId(id)
	body.SetType(typ)

	resp.SetLookup(&body)

	c.respond(&resp)
}

func (c *clientConn) respondCreate(id uint64) {
	var resp fsapi.Response

	var body fsapi.CreateResponse
	body.SetId(id)

	resp.SetCreate(&body)

	c.respond(&resp)
}

func (c *clientConn) respondReaddir(entries []*fsapi.Node) {
	var resp fsapi.Response

	var body fsapi.ReaddirResponse
	body.SetEntries(entries)

	resp.SetReaddir(&body)

	c.respond(&resp)
}

func (c *clientConn) respondGetattr(size uint64) {
	var resp fsapi.Response

	var body fsapi.GetattrResponse
	body.SetSize(size)

	resp.SetGetattr(&body)

	c.respond(&resp)
}

func (c *clientConn) respondWrite() {
	var resp fsapi.Response

	var body fsapi.WriteResponse

	resp.SetWrite(&body)

	c.respond(&resp)
}

func (c *clientConn) respondRead(data []byte) {
	var resp fsapi.Response

	var body fsapi.ReadResponse
	body.SetData(data)

	resp.SetRead(&body)

	c.respond(&resp)
}
