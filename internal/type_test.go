package internal

import (
	"errors"
	"fmt"
	"github.com/rpccloud/rpc/internal/util"
	"io/ioutil"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestRTPosRecord(t *testing.T) {
	assert := util.NewAssert(t)

	posArray := []int64{
		0,
		1,
		1000,
		0x00FFFFFFFFFFFFFE,
		0x00FFFFFFFFFFFFFF,
		0x7FFFFFFFFFFFFFFE,
		0x7FFFFFFFFFFFFFFF,
	}

	flagArray := []bool{
		false,
		true,
	}

	for i := 0; i < len(posArray); i++ {
		for j := 0; j < len(flagArray); j++ {
			pos := posArray[i]
			flag := flagArray[j]

			record := makePosRecord(pos, flag)

			fmt.Println("record", record)
			assert(record.getPos()).Equals(pos)
			assert(record.isString()).Equals(flag)

			fmt.Println(record)
		}
	}
}

func TestGetFuncKind(t *testing.T) {
	assert := util.NewAssert(t)

	fn1 := 3
	assert(getFuncKind(reflect.ValueOf(fn1))).
		Equals("", errors.New("handler must be a function"))

	fn2 := func() {}
	assert(getFuncKind(reflect.ValueOf(fn2))).
		Equals("", errors.New("handler 1st argument type must be rpc.Runtime"))

	fn3 := func(_ chan bool) {}
	assert(getFuncKind(reflect.ValueOf(fn3))).
		Equals("", errors.New("handler 1st argument type must be rpc.Runtime"))

	fn4 := func(rt Runtime, _ bool) {}
	assert(getFuncKind(reflect.ValueOf(fn4))).
		Equals("", errors.New("handler return type must be rpc.Return"))

	fn5 := func(rt Runtime, _ bool) (Return, bool) { return emptyReturn, true }
	assert(getFuncKind(reflect.ValueOf(fn5))).
		Equals("", errors.New("handler return type must be rpc.Return"))

	fn6 := func(rt Runtime, _ bool) bool { return true }
	assert(getFuncKind(reflect.ValueOf(fn6))).
		Equals("", errors.New("handler return type must be rpc.Return"))

	fn7 := func(rt Runtime,
		_ bool, _ int64, _ uint64, _ float64,
		_ string, _ Bytes, _ Array, _ Map,
	) Return {
		return rt.OK(true)
	}
	assert(getFuncKind(reflect.ValueOf(fn7))).Equals("BIUFSXAM", nil)

	fn8 := func(rt Runtime,
		_ int32, _ int64, _ uint64, _ float64,
		_ string, _ Bytes, _ Array, _ Map,
	) Return {
		return rt.OK(true)
	}
	assert(getFuncKind(reflect.ValueOf(fn8))).
		Equals("", errors.New("handler 2nd argument type int32 is not supported"))

	fn9 := func(rt Runtime,
		_ bool, _ int32, _ uint64, _ float64,
		_ string, _ Bytes, _ Array, _ Map,
	) Return {
		return rt.OK(true)
	}
	assert(getFuncKind(reflect.ValueOf(fn9))).
		Equals("", errors.New("handler 3rd argument type int32 is not supported"))

	fn10 := func(rt Runtime,
		_ bool, _ int64, _ int32, _ float64,
		_ string, _ Bytes, _ Array, _ Map,
	) Return {
		return rt.OK(true)
	}
	assert(getFuncKind(reflect.ValueOf(fn10))).
		Equals("", errors.New("handler 4th argument type int32 is not supported"))

	fn11 := func(rt Runtime,
		_ bool, _ int64, _ uint64, _ int32,
		_ string, _ Bytes, _ Array, _ Map,
	) Return {
		return rt.OK(true)
	}
	assert(getFuncKind(reflect.ValueOf(fn11))).
		Equals("", errors.New("handler 5th argument type int32 is not supported"))

	fn12 := func(rt Runtime,
		_ bool, _ int64, _ uint64, _ float64,
		_ int32, _ Bytes, _ Array, _ Map,
	) Return {
		return rt.OK(true)
	}
	assert(getFuncKind(reflect.ValueOf(fn12))).
		Equals("", errors.New("handler 6th argument type int32 is not supported"))

	fn13 := func(rt Runtime,
		_ bool, _ int64, _ uint64, _ float64,
		_ string, _ int32, _ Array, _ Map,
	) Return {
		return rt.OK(true)
	}
	assert(getFuncKind(reflect.ValueOf(fn13))).
		Equals("", errors.New("handler 7th argument type int32 is not supported"))

	fn14 := func(rt Runtime,
		_ bool, _ int64, _ uint64, _ float64,
		_ string, _ Bytes, _ int32, _ Map,
	) Return {
		return rt.OK(true)
	}
	assert(getFuncKind(reflect.ValueOf(fn14))).
		Equals("", errors.New("handler 8th argument type int32 is not supported"))

	fn15 := func(rt Runtime,
		_ bool, _ int64, _ uint64, _ float64,
		_ string, _ Bytes, _ Array, _ int32,
	) Return {
		return rt.OK(true)
	}
	assert(getFuncKind(reflect.ValueOf(fn15))).
		Equals("", errors.New("handler 9th argument type int32 is not supported"))
}

func TestConvertTypeToString(t *testing.T) {
	assert := util.NewAssert(t)
	assert(convertTypeToString(nil)).Equals("<nil>")
	assert(convertTypeToString(bytesType)).Equals("rpc.Bytes")
	assert(convertTypeToString(arrayType)).Equals("rpc.Array")
	assert(convertTypeToString(rtArrayType)).Equals("rpc.RTArray")
	assert(convertTypeToString(mapType)).Equals("rpc.Map")
	assert(convertTypeToString(rtMapType)).Equals("rpc.RTMap")
	assert(convertTypeToString(boolType)).Equals("rpc.Bool")
	assert(convertTypeToString(int64Type)).Equals("rpc.Int64")
	assert(convertTypeToString(uint64Type)).Equals("rpc.Uint64")
	assert(convertTypeToString(float64Type)).Equals("rpc.Float64")
	assert(convertTypeToString(stringType)).Equals("rpc.String")
	assert(convertTypeToString(contextType)).Equals("rpc.Runtime")
	assert(convertTypeToString(returnType)).Equals("rpc.Return")
	assert(convertTypeToString(reflect.ValueOf(make(chan bool)).Type())).
		Equals("chan bool")
}

func BenchmarkRTPosRecord(b *testing.B) {
	pos := int64(128)
	flag := true
	record := posRecord(0)

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		record = makePosRecord(pos, flag)
		pos = record.getPos()
		flag = record.isString()
	}

	fmt.Println(record)
}

func BenchmarkAddPrefixPerLine(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		util.AddPrefixPerLine("a\nb\nc", "test")
	}
}

func BenchmarkConcatString(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		util.ConcatString("a", "b")
	}
}

func BenchmarkGetCodePosition(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			util.AddFileLine("test", 0)
		}
	})
}

func BenchmarkGetRandString(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		util.GetRandString(128)
	}
}

func BenchmarkTimeNow(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		util.TimeNow()
	}
}

func BenchmarkTimeNowISOString(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		util.TimeNowISOString()
	}
}

func BenchmarkRunWithPanicCatch(b *testing.B) {
	a := uint64(0)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			testRunWithCatchPanic(func() {
				a = a + 1
			})
		}
	})
}

type testFuncCache struct{}

func (p *testFuncCache) Get(fnString string) ReplyCacheFunc {
	switch fnString {
	case "":
		return func(rt Runtime, stream *Stream, fn interface{}) bool {
			if !stream.IsReadFinish() {
				return false
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(Runtime) Return)(rt)
				return true
			}
		}
	case "S":
		return func(rt Runtime, stream *Stream, fn interface{}) bool {
			if arg0, ok := stream.ReadString(); !ok {
				return false
			} else if !stream.IsReadFinish() {
				return false
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(Runtime, String) Return)(rt, arg0)
				return true
			}
		}
	case "I":
		return func(rt Runtime, stream *Stream, fn interface{}) bool {
			if arg0, ok := stream.ReadInt64(); !ok {
				return false
			} else if !stream.IsReadFinish() {
				return false
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(Runtime, Int64) Return)(rt, arg0)
				return true
			}
		}
	case "M":
		return func(rt Runtime, stream *Stream, fn interface{}) bool {
			if arg0, ok := stream.ReadMap(); !ok {
				return false
			} else if !stream.IsReadFinish() {
				return false
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(Runtime, Map) Return)(rt, arg0)
				return true
			}
		}
	case "Y":
		return func(rt Runtime, stream *Stream, fn interface{}) bool {
			if arg0, ok := stream.ReadRTArray(rt); !ok {
				return false
			} else if !stream.IsReadFinish() {
				return false
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(Runtime, RTArray) Return)(rt, arg0)
				return true
			}
		}
	case "Z":
		return func(rt Runtime, stream *Stream, fn interface{}) bool {
			if arg0, ok := stream.ReadRTMap(rt); !ok {
				return false
			} else if !stream.IsReadFinish() {
				return false
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(Runtime, RTMap) Return)(rt, arg0)
				return true
			}
		}
	case "BIUFSXAM":
		return func(rt Runtime, stream *Stream, fn interface{}) bool {
			if arg0, ok := stream.ReadBool(); !ok {
				return false
			} else if arg1, ok := stream.ReadInt64(); !ok {
				return false
			} else if arg2, ok := stream.ReadUint64(); !ok {
				return false
			} else if arg3, ok := stream.ReadFloat64(); !ok {
				return false
			} else if arg4, ok := stream.ReadString(); !ok {
				return false
			} else if arg5, ok := stream.ReadBytes(); !ok {
				return false
			} else if arg6, ok := stream.ReadArray(); !ok {
				return false
			} else if arg7, ok := stream.ReadMap(); !ok {
				return false
			} else if !stream.IsReadFinish() {
				return false
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(
					Runtime, Bool, Int64, Uint64,
					Float64, String, Bytes, Array, Map,
				) Return)(rt, arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)
				return true
			}
		}
	default:
		return nil
	}
}

func testReadFromFile(filePath string) (string, Error) {
	ret, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", NewKernelPanic(err.Error())
	}

	// for windows, remove \r
	return strings.Replace(string(ret), "\r", "", -1), nil
}

func getFakeOnEvalBack() func(*Stream) {
	return func(stream *Stream) {}
}

func getFakeOnEvalFinish() func(*rpcThread) {
	return func(thread *rpcThread) {}
}

func getFakeProcessor(debug bool) *Processor {
	processor := NewProcessor(
		debug,
		1024,
		32,
		32,
		nil,
		5*time.Second,
		nil,
		func(stream *Stream) {},
	)
	processor.Close()
	return processor
}

func getFakeThread(debug bool) *rpcThread {
	return newThread(
		getFakeProcessor(debug),
		5*time.Second,
		getFakeOnEvalBack(),
		getFakeOnEvalFinish(),
	)
}

func testRunWithSubscribePanic(fn func()) Error {
	ch := make(chan Error, 1)
	sub := subscribePanic(func(err Error) {
		ch <- err
	})
	defer sub.Close()

	fn()

	select {
	case err := <-ch:
		return err
	default:
		return nil
	}
}

func testRunWithCatchPanic(fn func()) (ret interface{}) {
	defer func() {
		ret = recover()
	}()

	fn()
	return
}

func testRunWithProcessor(
	isDebug bool,
	fnCache ReplyCache,
	handler interface{},
	getStream func(processor *Processor) *Stream,
	onTest func(processor *Processor),
) (ret interface{}, retError Error, retPanic Error) {
	helper := newTestProcessorReturnHelper()
	service := NewService().Reply("Eval", handler)

	if processor := NewProcessor(
		isDebug,
		1024,
		16,
		16,
		fnCache,
		5*time.Second,
		[]*ServiceMeta{{
			name:     "test",
			service:  service,
			fileLine: "",
		}},
		helper.GetFunction(),
	); processor == nil {
		panic("internal error")
	} else if inStream := getStream(processor); inStream == nil {
		panic("internal error")
	} else {
		processor.PutStream(inStream)
		if onTest != nil {
			onTest(processor)
		}

		helper.WaitForFirstStream()

		if !processor.Close() {
			panic("internal error")
		}

		retArray, errorArray, panicArray := helper.GetReturn()

		if len(retArray) > 1 || len(errorArray) > 1 || len(panicArray) > 1 {
			panic("internal error")
		}

		if len(retArray) == 1 {
			ret = retArray[0]
		}

		if len(errorArray) == 1 {
			retError = errorArray[0]
		}

		if len(panicArray) == 1 {
			retPanic = panicArray[0]
		}

		return
	}
}

func testRunOnContext(
	isDebug bool,
	fn func(processor *Processor, rt Runtime) Return,
) (interface{}, Error, Error) {
	processorCH := make(chan *Processor)
	return testRunWithProcessor(
		isDebug,
		nil,
		func(rt Runtime) Return {
			return fn(<-processorCH, rt)
		},
		func(processor *Processor) *Stream {
			stream := NewStream()
			stream.SetDepth(3)
			stream.WriteString("#.test:Eval")
			stream.WriteString("")
			return stream
		},
		func(processor *Processor) {
			processorCH <- processor
		},
	)
}

type testProcessorReturnHelper struct {
	streamCH       chan *Stream
	firstReceiveCH chan bool
	isFirst        int32
}

func newTestProcessorReturnHelper() *testProcessorReturnHelper {
	return &testProcessorReturnHelper{
		streamCH:       make(chan *Stream, 102400),
		firstReceiveCH: make(chan bool, 1),
		isFirst:        0,
	}
}

func (p *testProcessorReturnHelper) GetFunction() func(stream *Stream) {
	return func(stream *Stream) {
		if atomic.CompareAndSwapInt32(&p.isFirst, 0, 1) {
			p.firstReceiveCH <- true
		}

		stream.SetReadPosToBodyStart()
		if kind, ok := stream.ReadUint64(); ok {
			if kind == uint64(ErrorKindTransport) {
				panic("it makes onEvalFinish panic")
			}
		}

		select {
		case p.streamCH <- stream:
			return
		case <-time.After(time.Second):
			// prevent capture
			go func() {
				panic("streamCH is full")
			}()
		}
	}
}

func (p *testProcessorReturnHelper) WaitForFirstStream() {
	<-p.firstReceiveCH
}

func (p *testProcessorReturnHelper) GetReturn() ([]Any, []Error, []Error) {
	retArray := make([]Any, 0)
	errorArray := make([]Error, 0)
	panicArray := make([]Error, 0)
	reportPanic := func(message string) {
		go func() {
			panic("message")
		}()
	}
	close(p.streamCH)
	for stream := range p.streamCH {
		stream.SetReadPosToBodyStart()
		if kind, ok := stream.ReadUint64(); !ok {
			reportPanic("stream is bad")
		} else if ErrorKind(kind) == ErrorKindNone {
			if v, ok := stream.Read(); ok {
				retArray = append(retArray, v)
			} else {
				reportPanic("read value error")
			}
		} else {
			if message, ok := stream.ReadString(); !ok {
				reportPanic("read message error")
			} else if debug, ok := stream.ReadString(); !ok {
				reportPanic("read debug error")
			} else {
				err := NewError(ErrorKind(kind), message, debug)

				switch ErrorKind(kind) {
				case ErrorKindProtocol:
					fallthrough
				case ErrorKindTransport:
					fallthrough
				case ErrorKindReply:
					errorArray = append(errorArray, err)
				case ErrorKindReplyPanic:
					fallthrough
				case ErrorKindRuntimePanic:
					fallthrough
				case ErrorKindKernelPanic:
					panicArray = append(panicArray, err)
				default:
					reportPanic("kind error")
				}
			}
		}
		stream.Release()
	}
	return retArray, errorArray, panicArray
}
