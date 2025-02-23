package storageserver

import (
	"github.com/overmighty/glan/glanfs/api/storageapi"
	"log/slog"
)

func (s *StorageServer) serve() {
	defer s.conn.Close()

	for {
		var req storageapi.Request
		err := s.conn.ReadMessage(&req)
		if err != nil {
			slog.Error("Failed to read message", "err", err, "remote_addr", s.conn.RemoteAddr())
			return
		}

		switch req.WhichBody() {
		case storageapi.Request_GetCapacity_case:
			s.handleGetCapacity()

		case storageapi.Request_Write_case:
			s.handleWrite(&req)
		case storageapi.Request_Read_case:
			s.handleRead(&req)

		default:
			slog.Error("Received invalid request from master server", "remote_addr", s.conn.RemoteAddr())
			return
		}
	}
}

func (s *StorageServer) handleGetCapacity() {
	s.respondGetCapacity()
}

func (s *StorageServer) handleWrite(req *storageapi.Request) {
	write := req.GetWrite()

	s.storage.write(write.GetBlockId(), write.GetData(), write.GetOffset())
	s.respondWrite()
}

func (s *StorageServer) handleRead(req *storageapi.Request) {
	read := req.GetRead()

	data := s.storage.read(read.GetBlockId(), read.GetSize(), read.GetOffset())
	s.respondRead(data)
}
