package ecs

import (
	"testing"
)

// goos: linux
// goarch: amd64
// pkg: github.com/unitoftime/ecs
// cpu: Intel(R) Core(TM) i7-8700K CPU @ 3.70GHz
// BenchmarkSetup-12                  	       2	 763499410 ns/op
// BenchmarkReads-12                  	       3	 354587956 ns/op
// BenchmarkWriteExisting-12          	       3	 390596487 ns/op
// BenchmarkWriteAdd-12               	       1	2715963517 ns/op
// BenchmarkLoopConstAdd-12           	     175	   6891824 ns/op
// BenchmarkLoopAdd-12                	     100	  11029501 ns/op
// BenchmarkLoopCompare-12            	     100	  10844822 ns/op
// BenchmarkLoopCompareReadOnly-12    	      31	  34887657 ns/op
// BenchmarkBaseline-12               	     854	   1460540 ns/op
// BenchmarkBaselinePointerMap-12     	     765	   1420869 ns/op
// BenchmarkBaselineMap-12            	     838	   1432447 ns/op

func setup(size int) *World {
	world := NewWorld()

	for i := 0; i < size; i++ {
		id := world.NewId()
		Write(world, id, d1{i}, d2{i})
	}
	return world
}


func BenchmarkSetup(b *testing.B) {
	for i := 0; i < b.N; i++ {
		setup(1e6)
	}
}

func BenchmarkReads(b *testing.B) {
	world := setup(1e6)
	b.ResetTimer()

	idStart := UniqueEntity + 1

	for i := 0; i < b.N; i++ {
		data := d1{}
		for j := idStart; j < 1e6+idStart; j++ {
			Read(world, Id(j), &data)
		}
	}
}

func BenchmarkWriteExisting(b *testing.B) {
	world := setup(1e6)
	b.ResetTimer()

	idStart := UniqueEntity + 1

	for i := 0; i < b.N; i++ {
		data := d1{0}
		for j := idStart; j < 1e6+idStart; j++ {
			Write(world, Id(j), data)
		}
	}
}

func BenchmarkWriteAdd(b *testing.B) {
	world := setup(1e6)
	b.ResetTimer()

	idStart := UniqueEntity + 1

	for i := 0; i < b.N; i++ {
		data := d3{0}
		for j := idStart; j < 1e6+idStart; j++ {
			Write(world, Id(j), data)
		}
	}
}

func BenchmarkLoopConstAdd(b *testing.B) {
	world := setup(1e6)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		view := ViewAll(world, &d1{})
		view.Map(func(id Id, comp ...interface{}) {
			data := comp[0].(*d1)
			data.value += 1
		})
	}
}

func BenchmarkLoopAdd(b *testing.B) {
	world := setup(1e6)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		view := ViewAll(world, &d1{}, &d2{})
		view.Map(func(id Id, comp ...interface{}) {
			data := comp[0].(*d1)
			data2 := comp[1].(*d2)
			data.value += data2.value
		})
	}
}

func BenchmarkLoopCompare(b *testing.B) {
	world := setup(1e6)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		view := ViewAll(world, &d1{}, &d2{})
		view.Map(func(id Id, comp ...interface{}) {
			data := comp[0].(*d1)
			data2 := comp[1].(*d2)
			if data.value != data2.value {
				b.Errorf("values should always match!")
			}
		})
	}
}

func BenchmarkLoopCompareReadOnly(b *testing.B) {
	world := setup(1e6)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		view := ViewAll(world, d1{}, d2{})
		view.Map(func(id Id, comp ...interface{}) {
			data := comp[0].(d1)
			data2 := comp[1].(d2)
			if data.value != data2.value {
				b.Errorf("values should always match!")
			}
		})
	}
}

// ---
// - Baseline Arrays
// ---
func BenchmarkBaseline(b *testing.B) {
	aa := make([]d1, 1e6)
	bb := make([]d2, 1e6)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for j := 0; j < 1e6; j++ {
			if aa[j].value != bb[j].value {
				b.Errorf("values should always match!")
			}
		}
	}
}

func pointerMapFunc(aa []d1, bb []d2, f func(index int, x *d1, y *d2)) {
	for j := 0; j < 1e6; j++ {
		f(j, &aa[j], &bb[j])
	}
}

func BenchmarkBaselinePointerMap(b *testing.B) {
	aa := make([]d1, 1e6)
	bb := make([]d2, 1e6)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pointerMapFunc(aa, bb, func(index int, x *d1, y *d2) {
			if x.value != y.value {
				b.Errorf("values should always match!")
			}
		})
	}
}

func mapFunc(aa []d1, bb []d2, f func(index int, x d1, y d2)) {
	for j := 0; j < 1e6; j++ {
		f(j, aa[j], bb[j])
	}
}

func BenchmarkBaselineMap(b *testing.B) {
	aa := make([]d1, 1e6)
	bb := make([]d2, 1e6)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mapFunc(aa, bb, func(index int, x d1, y d2) {
			if x.value != y.value {
				b.Errorf("values should always match!")
			}
		})
	}
}
