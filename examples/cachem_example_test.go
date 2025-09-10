package examples

import (
	"testing"

	"github.com/sunvim/utils/cachem"
)

func TestCachemExample(t *testing.T) {
	t.Log("Testing cachem (memory cache) package")

	// 测试内存分配和释放
	buf := cachem.Malloc(1024) // 分配1KB内存
	if len(buf) != 1024 {
		t.Fatalf("Expected length 1024, got %d", len(buf))
	}

	// 写入数据到缓冲区
	for i := range buf {
		buf[i] = byte(i % 256)
	}

	// 验证数据
	for i, b := range buf {
		if b != byte(i%256) {
			t.Fatalf("Data mismatch at index %d", i)
		}
	}

	// 释放内存
	cachem.Free(buf)

	t.Log("cachem 测试通过 - 内存分配和释放正常工作")

	// 测试带容量参数的分配
	buf2 := cachem.Malloc(512, 1024) // 长度512，容量至少1024
	if len(buf2) != 512 {
		t.Fatalf("Expected length 512, got %d", len(buf2))
	}
	if cap(buf2) < 1024 {
		t.Fatalf("Expected capacity >= 1024, got %d", cap(buf2))
	}

	cachem.Free(buf2)
	t.Log("cachem 带容量参数的分配测试通过")
}

func BenchmarkCachemMalloc(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := cachem.Malloc(1024)
		cachem.Free(buf)
	}
}