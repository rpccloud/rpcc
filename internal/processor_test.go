package internal

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewProcessor(t *testing.T) {
	assert := NewAssert(t)

	// Test(1) onReturnStream is nil
	assert(NewProcessor(true, 1, 1, 1, nil, 5*time.Second, nil, nil)).IsNil()

	// Test(2) numOfThreads <= 0
	helper2 := newTestProcessorReturnHelper()
	assert(
		NewProcessor(true, 0, 1, 1, nil, 5*time.Second, nil, helper2.GetFunction()),
	).IsNil()
	_, _, panicArray2 := helper2.GetReturn()
	assert(len(panicArray2)).Equals(1)
	assert(panicArray2[0].GetKind()).Equals(ErrorKindKernelPanic)
	assert(panicArray2[0].GetMessage()).Equals("rpc: numOfThreads is wrong")
	assert(strings.Contains(panicArray2[0].GetDebug(), "NewProcessor")).IsTrue()

	// Test(3) maxNodeDepth <= 0
	helper3 := newTestProcessorReturnHelper()
	assert(
		NewProcessor(true, 1, 0, 1, nil, 5*time.Second, nil, helper3.GetFunction()),
	).IsNil()
	_, _, panicArray3 := helper3.GetReturn()
	assert(len(panicArray3)).Equals(1)
	assert(panicArray3[0].GetKind()).Equals(ErrorKindKernelPanic)
	assert(panicArray3[0].GetMessage()).Equals("rpc: maxNodeDepth is wrong")
	assert(strings.Contains(panicArray3[0].GetDebug(), "NewProcessor")).IsTrue()

	// Test(4) maxCallDepth <= 0
	helper4 := newTestProcessorReturnHelper()
	assert(
		NewProcessor(true, 1, 1, 0, nil, 5*time.Second, nil, helper4.GetFunction()),
	).IsNil()
	_, _, panicArray4 := helper4.GetReturn()
	assert(len(panicArray4)).Equals(1)
	assert(panicArray4[0].GetKind()).Equals(ErrorKindKernelPanic)
	assert(panicArray4[0].GetMessage()).Equals("rpc: maxCallDepth is wrong")
	assert(strings.Contains(panicArray4[0].GetDebug(), "NewProcessor")).IsTrue()

	// Test(5) mount service error
	helper5 := newTestProcessorReturnHelper()
	assert(NewProcessor(
		true,
		1,
		1,
		1,
		nil,
		5*time.Second,
		[]*ServiceMeta{nil},
		helper5.GetFunction(),
	)).IsNil()
	_, _, panicArray5 := helper5.GetReturn()
	assert(len(panicArray5)).Equals(1)
	assert(panicArray5[0].GetKind()).Equals(ErrorKindKernelPanic)
	assert(panicArray5[0].GetMessage()).Equals("rpc: nodeMeta is nil")
	assert(strings.Contains(panicArray5[0].GetDebug(), "NewProcessor")).IsTrue()

	// Test(6) OK
	helper6 := newTestProcessorReturnHelper()
	processor6 := NewProcessor(
		true,
		65535,
		2,
		3,
		nil,
		5*time.Second,
		[]*ServiceMeta{{
			name: "test",
			service: NewService().Reply("Eval", func(ctx Context) Return {
				time.Sleep(time.Second)
				return ctx.OK(true)
			}),
			fileLine: "",
		}},
		helper6.GetFunction(),
	)
	for i := 0; i < 65536; i++ {
		stream := NewStream()
		stream.WriteString("#.test:Eval")
		stream.WriteUint64(3)
		stream.WriteString("")
		processor6.PutStream(stream)
	}
	assert(processor6).IsNotNil()
	assert(processor6.isDebug).IsTrue()
	assert(len(processor6.repliesMap)).Equals(1)
	assert(len(processor6.servicesMap)).Equals(2)
	assert(processor6.maxNodeDepth).Equals(uint64(2))
	assert(processor6.maxCallDepth).Equals(uint64(3))
	assert(len(processor6.threads)).Equals(65536)
	assert(len(processor6.freeCHArray)).Equals(freeGroups)
	assert(processor6.readThreadPos).Equals(uint64(65536))
	assert(processor6.fnError).IsNotNil()
	processor6.Close()
	assert(processor6.writeThreadPos).Equals(uint64(65536))
	sumFrees := 0
	for _, freeCH := range processor6.freeCHArray {
		sumFrees += len(freeCH)
	}
	assert(sumFrees).Equals(65536)
}

func TestProcessor_Close(t *testing.T) {
	assert := NewAssert(t)

	// Test(1) p.panicSubscription == nil
	processor1 := getFakeProcessor(true)
	assert(processor1.Close()).IsFalse()

	// Test(2)
	replyFileLine2 := ""
	helper2 := newTestProcessorReturnHelper()
	processor2 := NewProcessor(
		true,
		1024,
		2,
		3,
		nil,
		time.Second,
		[]*ServiceMeta{{
			name: "test",
			service: NewService().Reply("Eval", func(ctx Context) Return {
				replyFileLine2 = ctx.getThread().GetExecReplyFileLine()
				time.Sleep(2 * time.Second)
				return ctx.OK(true)
			}),
			fileLine: "",
		}},
		helper2.GetFunction(),
	)
	for i := 0; i < 1; i++ {
		stream := NewStream()
		stream.WriteString("#.test:Eval")
		stream.WriteUint64(3)
		stream.WriteString("")
		processor2.PutStream(stream)
	}
	assert(processor2.Close()).IsFalse()
	assert(helper2.GetReturn()).Equals([]Any{}, []Error{}, []Error{
		NewReplyPanic(
			"rpc: the following replies can not close: \n\t" +
				replyFileLine2 + " (1 goroutine)",
		),
	})

	// Test(3)
	replyFileLine3 := ""
	helper3 := newTestProcessorReturnHelper()
	processor3 := NewProcessor(
		true,
		1024,
		2,
		3,
		nil,
		time.Second,
		[]*ServiceMeta{{
			name: "test",
			service: NewService().Reply("Eval", func(ctx Context) Return {
				replyFileLine3 = ctx.getThread().GetExecReplyFileLine()
				time.Sleep(2 * time.Second)
				return ctx.OK(true)
			}),
			fileLine: "",
		}},
		helper3.GetFunction(),
	)
	for i := 0; i < 2; i++ {
		stream := NewStream()
		stream.WriteString("#.test:Eval")
		stream.WriteUint64(3)
		stream.WriteString("")
		processor3.PutStream(stream)
	}
	assert(processor3.Close()).IsFalse()
	assert(helper3.GetReturn()).Equals([]Any{}, []Error{}, []Error{
		NewReplyPanic(
			"rpc: the following replies can not close: \n\t" +
				replyFileLine3 + " (2 goroutines)",
		),
	})
}

func TestProcessor_PutStream(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	processor1 := NewProcessor(
		true,
		1024,
		32,
		32,
		nil,
		5*time.Second,
		nil,
		func(stream *Stream) {},
	)
	defer processor1.Close()
	assert(processor1.PutStream(NewStream())).IsTrue()

	// Test(2)
	processor2 := getFakeProcessor(true)
	for i := 0; i < 2048; i++ {
		assert(processor2.PutStream(NewStream())).IsFalse()
	}
}

func TestProcessor_Panic(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	err1 := NewRuntimePanic("message").AddDebug("debug")
	helper1 := newTestProcessorReturnHelper()
	processor1 := NewProcessor(
		true,
		1,
		1,
		1,
		nil,
		5*time.Second,
		nil,
		helper1.GetFunction(),
	)
	defer processor1.Close()
	processor1.Panic(err1)
	assert(helper1.GetReturn()).Equals([]Any{}, []Error{}, []Error{err1})
}

func TestProcessor_BuildCache(t *testing.T) {
	assert := NewAssert(t)
	_, file, _, _ := runtime.Caller(0)
	currDir := path.Dir(file)
	defer func() {
		_ = os.RemoveAll(path.Join(path.Dir(file), "_tmp_"))
	}()

	// Test(1)
	helper3 := newTestProcessorReturnHelper()
	processor3 := NewProcessor(
		true,
		1024,
		2,
		3,
		nil,
		5*time.Second,
		nil,
		helper3.GetFunction(),
	)
	defer processor3.Close()
	assert(processor3.BuildCache(
		"pkgName",
		path.Join(currDir, "processor_test.go/err_dir.go"),
	)).IsFalse()
	_, _, panicArray3 := helper3.GetReturn()
	assert(len(panicArray3)).Equals(1)
	assert(panicArray3[0].GetKind()).Equals(ErrorKindRuntimePanic)
	assert(strings.Contains(panicArray3[0].GetMessage(), "processor_test.go")).
		IsTrue()

	// Test(2)
	tmpFile2 := path.Join(currDir, "_tmp_/test-processor-02.go")
	snapshotFile2 := path.Join(currDir, "snapshot/test-processor-02.snapshot")
	helper2 := newTestProcessorReturnHelper()
	processor2 := NewProcessor(
		true,
		1024,
		2,
		3,
		nil,
		5*time.Second,
		[]*ServiceMeta{{
			name: "test",
			service: NewService().Reply("Eval", func(ctx Context) Return {
				return ctx.OK(true)
			}),
			fileLine: "",
		}},
		helper2.GetFunction(),
	)
	defer processor2.Close()
	assert(processor2.BuildCache("pkgName", tmpFile2)).IsTrue()
	assert(helper2.GetReturn()).Equals([]Any{}, []Error{}, []Error{})
	assert(testReadFromFile(tmpFile2)).Equals(testReadFromFile(snapshotFile2))
}

func TestProcessor_mountNode(t *testing.T) {
	assert := NewAssert(t)

	fnTestMount := func(
		services []*ServiceMeta,
		wantPanicKind ErrorKind,
		wantPanicMessage string,
		wantPanicDebug string,
	) {
		helper := newTestProcessorReturnHelper()
		assert(NewProcessor(
			true,
			1024,
			2,
			3,
			&testFuncCache{},
			5*time.Second,
			services,
			helper.GetFunction(),
		)).IsNil()
		retArray, errArray, panicArray := helper.GetReturn()
		assert(retArray, errArray).Equals([]Any{}, []Error{})
		assert(len(panicArray)).Equals(1)

		assert(panicArray[0].GetKind()).Equals(wantPanicKind)
		assert(panicArray[0].GetMessage()).Equals(wantPanicMessage)

		if wantPanicKind == ErrorKindKernelPanic {
			assert(strings.Contains(panicArray[0].GetDebug(), "goroutine")).IsTrue()
			assert(strings.Contains(panicArray[0].GetDebug(), "[running]")).IsTrue()
			assert(strings.Contains(panicArray[0].GetDebug(), "mountNode")).IsTrue()
			assert(strings.Contains(panicArray[0].GetDebug(), "NewProcessor")).
				IsTrue()
		} else {
			assert(panicArray[0].GetDebug()).Equals(wantPanicDebug)
		}
	}

	// Test(1)
	fnTestMount([]*ServiceMeta{
		nil,
	}, ErrorKindKernelPanic, "rpc: nodeMeta is nil", "")

	// Test(2)
	fnTestMount([]*ServiceMeta{{
		name:     "+",
		service:  NewService(),
		fileLine: "DebugMessage",
	}}, ErrorKindRuntimePanic, "rpc: service name + is illegal", "DebugMessage")

	// Test(3)
	fnTestMount([]*ServiceMeta{{
		name:     "abc",
		service:  nil,
		fileLine: "DebugMessage",
	}}, ErrorKindRuntimePanic, "rpc: service is nil", "DebugMessage")

	// Test(4)
	s4, source1 := NewService().AddChildService("s", NewService()), GetFileLine(0)
	fnTestMount(
		[]*ServiceMeta{{
			name:     "s",
			service:  NewService().AddChildService("s", s4),
			fileLine: "DebugMessage",
		}},
		ErrorKindRuntimePanic,
		"rpc: service path #.s.s.s is too long",
		source1,
	)

	// Test(5)
	fnTestMount(
		[]*ServiceMeta{{
			name:     "user",
			service:  NewService(),
			fileLine: "Debug1",
		}, {
			name:     "user",
			service:  NewService(),
			fileLine: "Debug2",
		}},
		ErrorKindRuntimePanic,
		"rpc: duplicated service name user",
		"current:\n\tDebug2\nconflict:\n\tDebug1",
	)

	// Test(6)
	fnTestMount(
		[]*ServiceMeta{{
			name: "user",
			service: &Service{
				children: []*ServiceMeta{},
				replies:  []*rpcReplyMeta{nil},
				fileLine: "DebugReply",
			},
			fileLine: "DebugService",
		}},
		ErrorKindKernelPanic,
		"rpc: meta is nil",
		"DebugReply",
	)

	// Test(7)
	fnTestMount(
		[]*ServiceMeta{{
			name: "test",
			service: &Service{
				children: []*ServiceMeta{},
				replies: []*rpcReplyMeta{{
					name:     "-",
					handler:  nil,
					fileLine: "DebugReply",
				}},
				fileLine: "DebugService",
			},
			fileLine: "Debug1",
		}},
		ErrorKindRuntimePanic,
		"rpc: reply name - is illegal",
		"DebugReply",
	)

	// Test(8)
	fnTestMount(
		[]*ServiceMeta{{
			name: "test",
			service: &Service{
				children: []*ServiceMeta{},
				replies: []*rpcReplyMeta{{
					name:     "Eval",
					handler:  nil,
					fileLine: "DebugReply",
				}},
				fileLine: "DebugService",
			},
			fileLine: "Debug1",
		}},
		ErrorKindRuntimePanic,
		"rpc: handler is nil",
		"DebugReply",
	)

	// Test(9)
	fnTestMount(
		[]*ServiceMeta{{
			name: "test",
			service: &Service{
				children: []*ServiceMeta{},
				replies: []*rpcReplyMeta{{
					name:     "Eval",
					handler:  3,
					fileLine: "DebugReply",
				}},
				fileLine: "DebugService",
			},
			fileLine: "Debug1",
		}},
		ErrorKindRuntimePanic,
		"rpc: handler must be func(ctx rpc.Context, ...) rpc.Return",
		"DebugReply",
	)

	// Test(10)
	fnTestMount(
		[]*ServiceMeta{{
			name: "test",
			service: &Service{
				children: []*ServiceMeta{},
				replies: []*rpcReplyMeta{{
					name:     "Eval",
					handler:  func() {},
					fileLine: "DebugReply",
				}},
				fileLine: "DebugService",
			},
			fileLine: "Debug1",
		}},
		ErrorKindRuntimePanic,
		"handler 1st argument type must be rpc.Context",
		"DebugReply",
	)

	// Test(11)
	fnTestMount(
		[]*ServiceMeta{{
			name: "test",
			service: &Service{
				children: []*ServiceMeta{},
				replies: []*rpcReplyMeta{{
					name:     "Eval",
					handler:  func(ctx Context) Return { return ctx.OK(true) },
					fileLine: "DebugReply1",
				}, {
					name:     "Eval",
					handler:  func(ctx Context) Return { return ctx.OK(true) },
					fileLine: "DebugReply2",
				}},
				fileLine: "DebugService",
			},
			fileLine: "Debug1",
		}},
		ErrorKindRuntimePanic,
		"rpc: reply name Eval is duplicated",
		"current:\n\tDebugReply2\nconflict:\n\tDebugReply1",
	)

	// Test(12)
	replyMeta12Eval1 := &rpcReplyMeta{
		name: "Eval1",
		handler: func(ctx Context, _a Array) Return {
			return ctx.OK(true)
		},
		fileLine: "DebugEval1",
	}
	replyMeta12Eval2 := &rpcReplyMeta{
		name: "Eval2",
		handler: func(ctx Context,
			_ bool, _ int64, _ uint64, _ float64,
			_ string, _ Bytes, _ Array, _ Map,
		) Return {
			return ctx.OK(true)
		},
		fileLine: "DebugEval2",
	}
	addMeta12 := &ServiceMeta{
		name: "test",
		service: &Service{
			children: []*ServiceMeta{},
			replies:  []*rpcReplyMeta{replyMeta12Eval1, replyMeta12Eval2},
			fileLine: GetFileLine(1),
		},
		fileLine: "serviceDebug",
	}
	processor12 := NewProcessor(
		true,
		1024,
		2,
		3,
		&testFuncCache{},
		5*time.Second,
		[]*ServiceMeta{addMeta12},
		getFakeOnEvalBack(),
	)
	assert(processor12).IsNotNil()
	defer processor12.Close()
	assert(*processor12.servicesMap["#"]).Equals(rpcServiceNode{
		path:    rootName,
		addMeta: nil,
		depth:   0,
	})
	assert(*processor12.servicesMap["#.test"]).Equals(rpcServiceNode{
		path:    "#.test",
		addMeta: addMeta12,
		depth:   1,
	})
	assert(processor12.repliesMap["#.test:Eval1"].path).Equals("#.test:Eval1")
	assert(processor12.repliesMap["#.test:Eval1"].meta).Equals(replyMeta12Eval1)
	assert(processor12.repliesMap["#.test:Eval1"].cacheFN).IsNil()
	assert(getFuncKind(processor12.repliesMap["#.test:Eval1"].reflectFn)).
		Equals("A", nil)
	assert(processor12.repliesMap["#.test:Eval1"].callString).
		Equals("#.test:Eval1(rpc.Context, rpc.Array) rpc.Return")
	assert(processor12.repliesMap["#.test:Eval1"].argTypes).
		Equals([]reflect.Type{contextType, arrayType})
	assert(processor12.repliesMap["#.test:Eval1"].indicator).IsNotNil()

	assert(processor12.repliesMap["#.test:Eval2"].path).Equals("#.test:Eval2")
	assert(processor12.repliesMap["#.test:Eval2"].meta).Equals(replyMeta12Eval2)
	assert(processor12.repliesMap["#.test:Eval2"].cacheFN).IsNotNil()
	assert(getFuncKind(processor12.repliesMap["#.test:Eval2"].reflectFn)).
		Equals("BIUFSXAM", nil)
	assert(processor12.repliesMap["#.test:Eval2"].callString).Equals(
		"#.test:Eval2(rpc.Context, rpc.Bool, rpc.Int64, rpc.Uint64, " +
			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return",
	)
	assert(processor12.repliesMap["#.test:Eval2"].argTypes).
		Equals([]reflect.Type{
			contextType, boolType, int64Type, uint64Type,
			float64Type, stringType, bytesType, arrayType, mapType,
		})
	assert(processor12.repliesMap["#.test:Eval2"].indicator).IsNotNil()
}

func BenchmarkRpcProcessor_Execute(b *testing.B) {
	total := uint64(0)
	success := uint64(0)
	failed := uint64(0)
	processor := NewProcessor(
		false,
		8192*24,
		16,
		16,
		&testFuncCache{},
		5*time.Second,
		[]*ServiceMeta{{
			name: "user",
			service: NewService().
				Reply("sayHello", func(ctx Context, name String) Return {
					return ctx.OK(name)
				}),
			fileLine: "",
		}},
		func(stream *Stream) {
			stream.SetReadPosToBodyStart()

			if kind, ok := stream.ReadUint64(); ok && kind == uint64(ErrorKindNone) {
				atomic.AddUint64(&success, 1)
			} else {
				atomic.AddUint64(&failed, 1)
			}

			stream.Release()
		},
	)

	b.ReportAllocs()
	b.N = 100000000
	b.SetParallelism(1024)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			stream := NewStream()
			stream.WriteString("#.user:sayHello")
			stream.WriteUint64(3)
			stream.WriteString("")
			stream.WriteString("")
			atomic.AddUint64(&total, 1)
			processor.PutStream(stream)
		}
	})

	fmt.Println(processor.Close())
	fmt.Println(total, success, failed)
}
