package ecs

import (
	"testing"
)

// goos: linux
// goarch: amd64
// pkg: github.com/jstewart7/ecs
// cpu: Intel(R) Core(TM) i7-8700K CPU @ 3.70GHz
// BenchmarkSetup-12            	       2	 792977374 ns/op
// BenchmarkReads-12            	       3	 368598063 ns/op
// BenchmarkWriteExisting-12    	       3	 395403302 ns/op
// BenchmarkWriteAdd-12         	       1	2700787354 ns/op
// BenchmarkLoopConstAdd-12     	     184	   6633520 ns/op
// BenchmarkLoopAdd-12          	     122	   9680233 ns/op
// BenchmarkLoopCompare-12      	     121	   9915709 ns/op

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
