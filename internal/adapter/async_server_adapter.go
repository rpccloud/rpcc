package adapter

import (
	"crypto/tls"
	"net"
	"reflect"
	"runtime"

	"github.com/rpccloud/rpc/internal/adapter/netpoll"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
)

// func websocketFD(conn net.Conn) int {
// 	tcpConn := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
// 	fdVal := tcpConn.FieldByName("fd")
// 	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")

// 	return int(pfdVal.FieldByName("Sysfd").Int())
// }

func parseFD(conn net.Conn, isTLS bool) int {
	if !isTLS {
		c := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
		fdVal := c.FieldByName("fd")
		pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")
		return int(pfdVal.FieldByName("Sysfd").Int())
	} else {
		tcpConn := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
		c := reflect.Indirect(tcpConn.Elem()).FieldByName("conn")
		fdVal := c.FieldByName("fd")
		pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")
		return int(pfdVal.FieldByName("Sysfd").Int())
	}
}

// AsyncServerAdapter ...
type AsyncServerAdapter struct {
	network   string
	addr      string
	rBufSize  int
	wBufSize  int
	tlsConfig *tls.Config
	receiver  IReceiver
	manager   *netpoll.Manager
}

// NewAsyncServerAdapter ...
func NewAsyncServerAdapter(
	network string,
	addr string,
	tlsConfig *tls.Config,
	rBufSize int,
	wBufSize int,
	receiver IReceiver,
) *RunnableService {

	return NewRunnableService(&AsyncServerAdapter{
		network:   network,
		addr:      addr,
		tlsConfig: tlsConfig,
		rBufSize:  rBufSize,
		wBufSize:  wBufSize,
		receiver:  receiver,
	})
}

// OnRun ...
func (p *AsyncServerAdapter) OnRun(service *RunnableService) {
	manager := netpoll.NewManager(
		func(err *base.Error) {
			p.receiver.OnConnError(nil, err)
		},
		base.MaxInt(runtime.NumCPU()/2, 1),
	)

	if manager == nil {
		return
	}
	defer manager.Close()

	switch p.network {
	case "tcp":
		p.runAsTCPServer(manager, service)
	default:
		panic("not implemented")
	}
}

func (p *AsyncServerAdapter) runAsTCPServer(
	manager *netpoll.Manager,
	service *RunnableService,
) {
	listener := net.Listener(nil)
	e := error(nil)

	if p.tlsConfig == nil {
		listener, e = net.Listen(p.network, p.addr)
	} else {
		listener, e = tls.Listen(p.network, p.addr, p.tlsConfig)
	}

	if e != nil {
		p.receiver.OnConnError(nil, errors.ErrTemp.AddDebug(e.Error()))
		return
	}

	for service.IsRunning() {
		tcpConn, e := listener.Accept()

		if e != nil {
			p.receiver.OnConnError(nil, errors.ErrTemp.AddDebug(e.Error()))
		} else {
			conn := NewNetConn(tcpConn, p.rBufSize, p.wBufSize)
			conn.SetNext(NewStreamConn(conn, p.receiver))
			conn.SetFD(parseFD(tcpConn, p.tlsConfig != nil))
			manager.AllocChannel().AddConn(conn)
		}
	}

	if e = listener.Close(); e != nil {
		p.receiver.OnConnError(nil, errors.ErrTemp.AddDebug(e.Error()))
	}

	listener.Close()
}

// OnStop ...
func (p *AsyncServerAdapter) OnStop(_ *RunnableService) {

}

// Close ...
func (p *AsyncServerAdapter) Close() {
	// do nothing
}
