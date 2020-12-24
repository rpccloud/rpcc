package main

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/adapter/common"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"time"
)

type receiver struct {
	streamCH   chan *core.Stream
	streamConn *common.StreamConn
}

func (p *receiver) OnConnOpen(streamConn *common.StreamConn) {
	fmt.Println("Client: OnConnOpen")
	p.streamConn = streamConn
}

func (p *receiver) OnConnClose(streamConn *common.StreamConn) {
	fmt.Println("Client: OnConnClose")

	p.streamConn = nil
}

func (p *receiver) OnConnReadStream(
	streamConn *common.StreamConn,
	stream *core.Stream,
) {
	p.streamCH <- stream
}

func (p *receiver) OnConnError(
	streamConn *common.StreamConn,
	err *base.Error,
) {
	if streamConn != nil {
		streamConn.Close()
	}

	fmt.Println("Client: OnConnError", err)
}

func main() {
	tlsConfig, err := base.GetTLSClientConfig(true, []string{"../cert/ca.pem"})
	if err != nil {
		panic(err)
	}

	clientReceiver := &receiver{streamCH: make(chan *core.Stream)}
	clientAdapter := adapter.NewClientAdapter(
		"tcp", "127.0.0.1:8080", tlsConfig, 1200, 1200, clientReceiver,
	)

	go func() {
		clientAdapter.Open()
		clientAdapter.Run()
	}()

	time.Sleep(time.Second)
	if clientReceiver.streamConn == nil {
		panic("not connect")
	}

	fmt.Println("Start Test", time.Now())
	start := time.Now()

	for i := 0; i < 1000000; i++ {
		stream := core.NewStream()
		stream.WriteInt64(12)
		stream.BuildStreamCheck()
		clientReceiver.streamConn.WriteStreamAndRelease(stream)
		s := <-clientReceiver.streamCH
		s.SetReadPosToBodyStart()
		if v, _ := s.ReadInt64(); v != int64(12) {
			panic("error")
		}

		if i%10000 == 0 {
			fmt.Println(i)
		}

		s.Release()
	}

	fmt.Println("End Test", time.Now().Sub(start))
	clientAdapter.Close()
	return
}
