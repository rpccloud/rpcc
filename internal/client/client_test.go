package client

import (
	"fmt"
	"testing"
	"time"

	"github.com/rpccloud/rpc/internal/server"
)

func TestClient_Debug(t *testing.T) {
	rpcServer := server.NewServer().Listen("tcp", "0.0.0.0:28888", nil)

	go func() {
		rpcServer.Serve()
	}()

	time.Sleep(3000 * time.Millisecond)

	rpcClient := newClient("tcp", "0.0.0.0:28888", nil, 1200, 1200)

	for i := 0; i < 2; i++ {
		fmt.Println(rpcClient.SendMessage(20*time.Second, "#.test:SayHello", i))
	}

	rpcServer.Close()
}
