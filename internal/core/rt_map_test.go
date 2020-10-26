package core

import (
	"fmt"
	"testing"
)

func TestRTMap_swapUint32(t *testing.T) {
	fmt.Println(sizeOfMapItem)
	rtMap := streamTestRuntime.NewRTMap()

	rtMap.appendValue("v1", 34234)
	rtMap.appendValue("v8", 34234)
	rtMap.appendValue("v7", 34234)
	rtMap.appendValue("a7", 34234)
	rtMap.appendValue("a8", 34234)
	rtMap.appendValue("v5", 34234)
	rtMap.appendValue("v2", 34234)
	rtMap.appendValue("v3", 34234)
	rtMap.appendValue("a2", 34234)
	rtMap.appendValue("v4", 34234)
	rtMap.appendValue("v6", 34234)
	rtMap.appendValue("a1", 34234)
	rtMap.appendValue("a4", 34234)
	rtMap.appendValue("a3", 34234)
	rtMap.appendValue("a5", 34234)
	rtMap.appendValue("a6", 34234)

	fmt.Println(rtMap.items)
}

func BenchmarkMakeRequestStream(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		streamTestRuntime.thread.rtStream.Reset()
		streamTestRuntime.thread.rootFrame.Reset()
		rtMap := streamTestRuntime.NewRTMap()
		rtMap.appendValue("v1", 34234)
		rtMap.appendValue("v8", 34234)
		rtMap.appendValue("v7", 34234)
		rtMap.appendValue("a7", 34234)
		rtMap.appendValue("a8", 34234)
		rtMap.appendValue("v5", 34234)
		rtMap.appendValue("v2", 34234)
		rtMap.appendValue("v3", 34234)
		rtMap.appendValue("a2", 34234)
		rtMap.appendValue("v4", 34234)
		rtMap.appendValue("v6", 34234)
		rtMap.appendValue("a1", 34234)
		rtMap.appendValue("a4", 34234)
		rtMap.appendValue("a3", 34234)
		rtMap.appendValue("a5", 34234)
		rtMap.appendValue("a6", 34234)
	}
}
