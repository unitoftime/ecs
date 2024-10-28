package ecs

import "testing"

// Reflect: BenchmarkName-12    	21372271	        56.14 ns/op	       0 B/op	       0 allocs/op
// Reflect: BenchmarkName-12    	21361663	        56.60 ns/op	       0 B/op	       0 allocs/op
// Unsafe:  BenchmarkName-12    	32242930	        37.01 ns/op	       0 B/op	       0 allocs/op
// Unsafe:  BenchmarkName-12    	31874323	        36.90 ns/op	       0 B/op	       0 allocs/op
func BenchmarkName(b *testing.B) {
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		C(position{})
		C(velocity{})
		C(acceleration{})
		C(radius{})
	}
}
