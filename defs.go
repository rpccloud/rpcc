package rpc

import (
	"github.com/rpccloud/rpc/internal"
	"time"
)

const configServerReadLimit = int64(1024 * 1024)
const configServerWriteLimit = int64(1024 * 1024)
const configReadTimeout = 10 * time.Second
const configWriteTimeout = 1 * time.Second

const SystemStreamKindInit = int64(1)
const SystemStreamKindInitBack = int64(2)
const SystemStreamKindRequestIds = int64(3)
const SystemStreamKindRequestIdsBack = int64(4)

type IStreamConnection interface {
	ReadStream(timeout time.Duration, readLimit int64) (Stream, Error)
	WriteStream(stream Stream, timeout time.Duration, writeLimit int64) Error
	Close() Error
}

type IAdapter interface {
	ConnectString() string
	IsRunning() bool
	Open(onConnRun func(IStreamConnection), onError func(Error)) bool
	Close(onError func(Error)) bool
}

// Stream ...
type Stream = internal.Stream

// Bool ...
type Bool = internal.Bool

// Int64 ...
type Int64 = internal.Int64

// Uint64 ...
type Uint64 = internal.Uint64

// Float64 ...
type Float64 = internal.Float64

// String ...
type String = internal.String

// Bytes ...
type Bytes = internal.Bytes

// Any common Any type
type Any = internal.Any

// Array common Array type
type Array = internal.Array

// Map common Map type
type Map = internal.Map

// Error ...
type Error = internal.Error

// ReturnObject ...
type Return = internal.Return

// ContextObject ...
type Context = internal.Context

// NewService ...
var NewService = internal.NewService

// ReplyCache ...
type ReplyCache = internal.ReplyCache

// ReplyCacheFunc ...
type ReplyCacheFunc = internal.ReplyCacheFunc
