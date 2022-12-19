package main

import (
	"os"

	"github.com/asmyasnikov/ydb-docker/internal/certs"
)

func main() {
	ydbGrpcTlsDataPath := os.Getenv("YDB_GRPC_TLS_DATA_PATH")
	if len(os.Args) > 1 {
		ydbGrpcTlsDataPath = os.Args[1]
	}
	certs := certs.New(ydbGrpcTlsDataPath)
	if err := certs.Persist(); err != nil {
		panic(err)
	}
}
