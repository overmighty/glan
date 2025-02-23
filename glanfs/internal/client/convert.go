package client

import (
	"github.com/overmighty/glan/glanfs/api/fsapi"
	"syscall"
)

func fsapiErrorToErrno(err fsapi.Error) syscall.Errno {
	switch err {
	case fsapi.Error_ERROR_ENOENT:
		return syscall.ENOENT
	case fsapi.Error_ERROR_ENOTDIR:
		return syscall.ENOTDIR
	case fsapi.Error_ERROR_EEXIST:
		return syscall.EEXIST
	case fsapi.Error_ERROR_ENOSPC:
		return syscall.ENOSPC
	default:
		panic("unreachable")
	}
}

func fsapiFileTypeToMode(fileType fsapi.FileType) uint32 {
	switch fileType {
	case fsapi.FileType_FILE_TYPE_REGULAR:
		return syscall.S_IFREG
	case fsapi.FileType_FILE_TYPE_DIRECTORY:
		return syscall.S_IFDIR
	default:
		panic("unreachable")
	}
}
