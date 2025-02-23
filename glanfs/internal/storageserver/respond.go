package storageserver

import (
	"github.com/overmighty/glan/glanfs/api/storageapi"
	"log/slog"
)

func (s *StorageServer) respond(resp *storageapi.Response) {
	if err := s.conn.WriteMessage(resp); err != nil {
		slog.Error("Failed to write message", "err", err, "remote_addr", s.conn.RemoteAddr())
		panic(err)
	}
}

func (s *StorageServer) respondGetCapacity() {
	var resp storageapi.Response

	var body storageapi.GetCapacityResponse
	body.SetNumBlocks(s.capacityBlocks())

	resp.SetGetCapacity(&body)

	s.respond(&resp)
}

func (s *StorageServer) respondWrite() {
	var resp storageapi.Response

	var body storageapi.WriteResponse

	resp.SetWrite(&body)

	s.respond(&resp)
}

func (s *StorageServer) respondRead(data []byte) {
	var resp storageapi.Response

	var body storageapi.ReadResponse
	body.SetData(data)

	resp.SetRead(&body)

	s.respond(&resp)
}
