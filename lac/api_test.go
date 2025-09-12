/*
 * Linear Allocator
 *
 * Improve the memory allocation and garbage collection performance.
 *
 * Copyright (C) 2020-2023 crazybie@github.com.
 * https://github.com/crazybie/linear_ac
 */

package lac

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

type EnumA int32

const (
	EnumVal1 EnumA = 1
	EnumVal2 EnumA = 2
)

type PbItem struct {
	Id      *int
	Price   *int
	Class   *int
	Name    *string
	Active  *bool
	EnumVal *EnumA
}

type PbData struct {
	Age   *int
	Items []*PbItem
	InUse *PbItem
}

func Test_Smoke(t *testing.T) {
	ac := acPool.Get()
	defer ac.Release()

	d := New[PbData](ac)
	d.Age = ac.Int(11)

	n := 3
	for i := 0; i < n; i++ {
		item := New[PbItem](ac)
		item.Id = ac.Int(i + 1)
		item.Active = ac.Bool(true)
		item.Price = ac.Int(100 + i)
		item.Class = ac.Int(3 + i)
		item.Name = ac.String("name")
		d.Items = Append(ac, d.Items, item)
	}

	if *d.Age != 11 {
		t.Errorf("age")
	}

	if len(d.Items) != int(n) {
		t.Errorf("item")
	}
	for i := 0; i < n; i++ {
		if *d.Items[i].Id != i+1 {
			t.Errorf("item.id")
		}
		if *d.Items[i].Price != i+100 {
			t.Errorf("item.price")
		}
		if *d.Items[i].Class != i+3 {
			t.Errorf("item.class")
		}
	}
}

func Test_Alignment(t *testing.T) {
	ac := acPool.Get()
	defer ac.Release()

	for i := 1; i < 1024; i++ {
		p := ac.alloc(i, false)
		if (uintptr(p) & uintptr(ptrSize-1)) != 0 {
			t.Fail()
		}
	}
}

func Test_String(t *testing.T) {
	ac := acPool.Get()
	defer ac.Release()

	type D struct {
		s [5]*string
	}
	d := New[D](ac)
	for i := range d.s {
		d.s[i] = ac.String(fmt.Sprintf("str%v", i))
		runtime.GC()
	}
	for i, p := range d.s {
		if *p != fmt.Sprintf("str%v", i) {
			t.Errorf("elem %v is gced", i)
		}
	}
}

func Test_NewMap(t *testing.T) {
	ac := acPool.Get()
	defer ac.Release()

	type D struct {
		m map[int]*int
	}
	data := [10]*D{}
	for i := 0; i < len(data); i++ {
		d := New[D](ac)
		d.m = NewMap[int, *int](ac, 0)
		d.m[1] = ac.Int(i)
		data[i] = d
		runtime.GC()
	}
	for i, d := range data {
		if *d.m[1] != i {
			t.Fail()
		}
	}
}

func Test_NewSlice(t *testing.T) {
	acPool.EnableDebugMode(true)
	ac := acPool.Get()
	defer ac.Release()

	s := make([]*int, 0)
	s = Append(ac, s, ac.Int(2))
	if len(s) != 1 && *s[0] != 2 {
		t.Fail()
	}

	s = NewSlice[*int](ac, 0, 32)
	s = Append(ac, s, ac.Int(1))
	if cap(s) != 32 || *s[0] != 1 {
		t.Fail()
	}

	intSlice := []int{}
	intSlice = Append(ac, intSlice, 11)
	if len(intSlice) != 1 || intSlice[0] != 11 {
		t.Fail()
	}

	byteSlice := []byte{}
	byteSlice = Append(ac, byteSlice, byte(11))
	if len(byteSlice) != 1 || byteSlice[0] != 11 {
		t.Fail()
	}

	type Data struct {
		d [2]uint64
	}
	structSlice := []Data{}
	d1 := uint64(0xcdcdefefcdcdefdc)
	d2 := uint64(0xcfcdefefcdcfffde)
	structSlice = Append(ac, structSlice, Data{d: [2]uint64{d1, d2}})
	if len(structSlice) != 1 || structSlice[0].d[0] != d1 || structSlice[0].d[1] != d2 {
		t.Fail()
	}

	f := func() []int {
		var r []int = NewSlice[int](ac, 0, 1)
		r = Append(ac, r, 1)
		return r
	}
	r := f()
	if len(r) != 1 {
		t.Errorf("return slice")
	}

	{
		var s []*PbItem
		s = Append(ac, s, nil)
		if len(s) != 1 || s[0] != nil {
			t.Errorf("nil")
		}
	}
}

func Test_NewFromRaw(b *testing.T) {
	var ac *Allocator

	for i := 0; i < 3; i++ {
		d := NewFrom(ac, &PbItem{
			Id:    ac.Int(1 + i),
			Class: ac.Int(2 + i),
			Price: ac.Int(3 + i),
			Name:  ac.String("test"),
		})

		if *d.Id != 1+i {
			b.Fail()
		}
		if *d.Class != 2+i {
			b.Fail()
		}
		if *d.Price != 3+i {
			b.Fail()
		}
		if *d.Name != "test" {
			b.Fail()
		}
	}
}

func Test_NewFrom(b *testing.T) {
	ac := acPool.Get()
	defer ac.Release()

	for i := 0; i < 3; i++ {
		d := NewFrom(ac, &PbItem{
			Id:    ac.Int(1 + i),
			Class: ac.Int(2 + i),
			Price: ac.Int(3 + i),
			Name:  ac.String("test"),
		})

		if *d.Id != 1+i {
			b.Fail()
		}
		if *d.Class != 2+i {
			b.Fail()
		}
		if *d.Price != 3+i {
			b.Fail()
		}
		if *d.Name != "test" {
			b.Fail()
		}
	}
}

func Test_BuildInAllocator(t *testing.T) {
	var ac *Allocator
	defer ac.Release()

	item := New[PbItem](ac)
	item.Id = ac.Int(11)
	if *item.Id != 11 {
		t.Fail()
	}
	id2 := 22
	item = NewFrom(ac, &PbItem{Id: &id2})
	if *item.Id != 22 {
		t.Fail()
	}
	s := NewSlice[*PbItem](ac, 0, 3)
	if cap(s) != 3 || len(s) != 0 {
		t.Fail()
	}
	s = Append(ac, s, item)
	if len(s) != 1 || *s[0].Id != 22 {
		t.Fail()
	}
	m := NewMap[int, string](ac, 0)
	m[1] = "test"
	if m[1] != "test" {
		t.Fail()
	}
	e := EnumVal1
	v := NewEnum(ac, e)
	if *v != e {
		t.Fail()
	}
}

func Test_Enum(t *testing.T) {
	ac := acPool.Get()
	defer ac.Release()

	e := EnumVal2
	v := NewEnum(ac, e)
	if *v != e {
		t.Fail()
	}
}

func Test_AttachExternal(b *testing.T) {
	ac := acPool.Get()
	defer ac.Release()

	const sz = 100

	type D struct {
		d  [sz]*int
		s  [sz]string
		ar [sz][]int
	}

	fail := false

	d := New[D](ac)
	for i := 0; i < len(d.d); i++ {
		// pointer
		if fail {
			d.d[i] = new(int)
		} else {
			d.d[i] = Attach(ac, new(int))
		}
		*d.d[i] = i

		// string
		if fail {
			d.s[i] = fmt.Sprintf("%d", i)
		} else {
			d.s[i] = Attach(ac, fmt.Sprintf("%d", i))
		}

		// slice
		if fail {
			d.ar[i] = []int{0, 1, 2, 3}
		} else {
			d.ar[i] = Attach(ac, []int{0, 1, 2, 3})
		}

		runtime.GC()
	}

	for i := 0; i < len(d.d); i++ {
		if *d.d[i] != i {
			b.Errorf("int should not be gced.")
		}
		if d.s[i] != fmt.Sprintf("%d", i) {
			b.Errorf("string should not be gced.")
		}
		if len(d.ar[i]) != 4 {
			b.Errorf("slice should not be gced.")
		}
		for idx, v := range d.ar[i] {
			if v != idx {
				b.Errorf("slice gced")
			}
		}
	}
}

func Test_Append(t *testing.T) {
	ac := acPool.Get()
	defer ac.Release()

	m := map[int][]int{}

	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			m[i] = Append(ac, m[i], j)
		}
	}

	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			if m[i][j] != j {
				t.Fail()
			}
		}
	}
}

func TestPointerSlice(t *testing.T) {
	ac := acPool.Get()
	defer ac.Release()

	ptrs := NewSlice[*PbItem](ac, 5, 5)
	for _, p := range ptrs {
		if p != nil {
			t.Fail()
		}
	}

	ptrs = []*PbItem{New[PbItem](ac)}
	ptrs = Append(ac, ptrs, New[PbItem](ac))
	l := len(ptrs)
	// should have empty space
	if l == cap(ptrs) {
		t.Fail()
	}
	// unused space must be zeroed
	for _, p := range ptrs[:cap(ptrs)][l+1:] {
		if p != nil {
			t.Fail()
		}
	}
}

func Test_SliceAppendStructValue(t *testing.T) {
	ac := acPool.Get()
	defer ac.Release()

	type S struct {
		a int
		b float32
		c string
	}

	var s []S
	s = Append(ac, s, S{1, 2, "3"})
	if s[0].a != 1 || s[0].b != 2 || s[0].c != "3" {
		t.Fail()
	}
}

func Test_NilAc(t *testing.T) {
	var ac *Allocator

	type S struct {
		v int
	}
	o := New[S](ac)
	if o == nil {
		t.Fail()
	}

	f := NewFrom(ac, &S{})
	if f == nil {
		t.Fail()
	}

	m := NewMap[int, int](ac, 1)
	if m == nil {
		t.Fail()
	}

	s := NewSlice[byte](ac, 10, 10)
	if cap(s) != 10 || len(s) != 10 {
		t.Fail()
	}

	e := NewEnum(ac, EnumVal2)
	if *e != EnumVal2 {
		t.Fail()
	}

	i := ac.Int(1)
	if *i != 1 {
		t.Fail()
	}

	ss := ac.String("ss")
	if *ss != "ss" {
		t.Fail()
	}

	ac.IncRef()
	ac.DecRef()
	ac.Release()
}

func Test_SliceWrongCap(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			panic(fmt.Errorf("should panic: out of range"))
		}
	}()
	ac := acPool.Get()
	defer ac.Release()
	NewSlice[byte](ac, 10, 0)
}

//go:linkname findObject runtime.findObject
func findObject(p, refBase, refOff uintptr) (base uintptr, s uintptr, objIndex uintptr)

// NOTE: must run without -race flag.
//
// Fix random crash:
// runtime: pointer 0xc000073ff8 to unused region of span span.base()=0xc000072000 span.limit=0xc000073fb0 span.state=1
// fatal error: found bad pointer in Go heap (incorrect use of unsafe or cgo?)
func TestSliceWbPanic(t *testing.T) {
	if os.Getenv("SimulateCrash") != "" {
		BugfixClearPointerInMem = false
	}

	type mSpanStateBox struct {
		s uint8
	}

	type mspan struct {
		next      *mspan
		prev      *mspan
		list      uintptr
		startAddr uintptr
		npages    uintptr

		manualFreeList uintptr
		freeindex      uintptr
		nelems         uintptr
		allocCache     uint64
		allocBits      uintptr
		gcmarkBits     uintptr

		sweepgen              uint32
		divMul                uint32
		allocCount            uint16
		spanclass             uint8
		state                 mSpanStateBox
		needzero              uint8
		allocCountBeforeCache uint16
		elemsize              uintptr
		limit                 uintptr
	}

	for j := 0; j < 1000; j++ {
		ac := acPool.Get()
		ac.Int(0) // ensure one chunk

		// write a special rubbish
		s := (*[2]uintptr)((*sliceHeader)(ac.curChunk).Data)
		_, ms, _ := findObject(uintptr(unsafe.Pointer(new(PbItem))), 0, 0)
		span := (*mspan)(unsafe.Pointer(ms))
		s[1] = span.limit + 1024*8 - 8

		// simulate the write barrier
		a := NewSlice[uintptr](ac, 1, 1)
		findObject((uintptr)(unsafe.Pointer(a[0])), 0, 0)

		ac.Release()
	}
}

func TestSameSpan(t *testing.T) {
	ac := acPool.Get()
	defer ac.Release()

	o := New[PbItem](ac)
	obj, span, idx := findObject(uintptr(unsafe.Pointer(o)), 0, 0)
	for i := 0; i < 100; i++ {
		p := New[PbItem](ac)
		obj2, span2, idx2 := findObject(uintptr(unsafe.Pointer(p)), 0, 0)
		if obj2 != obj {
			t.Errorf("obj")
		}
		if span != span2 {
			t.Errorf("span")
		}
		if idx != idx2 {
			t.Errorf("idx")
		}
	}
}

// NOTE: run with "-race".
func TestSharedAc_NoRace(t *testing.T) {
	ac := acPool.Get()
	wg := sync.WaitGroup{}
	wg.Add(10)

	for i := 0; i < 10; i++ {
		ac.IncRef()
		go func() {

			var item *PbItem
			for j := 0; j < 1000; j++ {
				item = New[PbItem](ac)
				item.Class = Attach(ac, new(int))
				*item.Class = j
				item.Id = ac.Int(j)
			}
			runtime.KeepAlive(item)

			ac.DecRef()
			wg.Done()
		}()
	}

	ac.DecRef()
	wg.Wait()
	if n := ac.refCnt.Load(); n != 1 {
		t.Errorf("ref cnt:%v", n)
	}
}

func TestReinitPool(t *testing.T) {
	acPool.EnableDebugMode(true)

	var running atomic.Bool
	running.Store(true)

	wg := sync.WaitGroup{}
	wg.Add(100)
	for j := 0; j < 100; j++ {

		go func() {

			for running.Load() {

				acPoolMu.RLock()
				ac := acPool.Get()
				acPoolMu.RUnlock()

				i := New[PbItem](ac)
				i.Class = ac.Int(100)
				i.Id = ac.Int(200)

				time.Sleep(time.Duration(1) * time.Millisecond)

				if *i.Class != 100 || *i.Id != 200 {
					t.Fail()
					break
				}

				runtime.KeepAlive(i)
				ac.DecRef()
			}

			wg.Done()
		}()
	}

	time.Sleep(time.Duration(1) * time.Microsecond)

	acPoolMu.Lock()
	acPool = NewAllocatorPool("test", nil, 10000, 64*1024, 32*1000, 64*1000)
	acPoolMu.Unlock()

	// force to recycle the old pool.
	runtime.GC()
	time.Sleep(time.Duration(10) * time.Microsecond)

	running.Store(false)
	wg.Wait()
}

//go:linkname atomicwb runtime.atomicwb
func atomicwb(ptr *unsafe.Pointer, new unsafe.Pointer)

func TestNoAlloc(t *testing.T) {
	ac := acPool.Get()
	defer ac.DecRef()

	ac.Int(1)

	noMalloc(func() {
		_ = New[PbItem](ac)
	})

	noMalloc(func() {
		_ = ac.Int(1)
	})

	noMalloc(func() {
		_ = ac.String("1")
	})

	noMalloc(func() {
		a := NewSlice[*PbItem](ac, 1, 1)
		a = Append(ac, a, New[PbItem](ac))
		a = Append(ac, a, New[PbItem](ac))
		runtime.KeepAlive(a)
	})
}

func TestFreeMarkedObj(t *testing.T) {
	if os.Getenv("SimulateCrash") != "" {
		BugfixCorruptOtherMem = false
		defer func() {
			BugfixCorruptOtherMem = true
		}()
	}

	ac := acPool.Get()
	defer ac.Release()

	for n := 0; n < 1000; n++ {

		// 1. exhaust the chunk
		//--------------------
		_ = NewSlice[byte](ac, 0, acPool.chunkPool.ChunkSize)

		// 2. alloc a zero size slice
		//--------------------
		items := NewSlice[*PbItem](ac, 0, 0)
		heapObj := new(PbData)

		// 3. force a write barrier flush
		//--------------------
		wbBufSize := 256
		for i := 0; i < wbBufSize; i++ {
			dst := &(*sliceHeader)(unsafe.Pointer(&heapObj.Items)).Data
			src := (*sliceHeader)(unsafe.Pointer(&items)).Data
			atomicwb(dst, src)
			heapObj.Items = items
		}

		// 4. force a gc sweep
		//--------------------
		runtime.GC()

		runtime.KeepAlive(heapObj)
		ac.reset()
	}
}
