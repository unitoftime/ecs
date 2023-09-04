package intmap

import "testing"

func TestMap64(t *testing.T) {
	type pairs [][2]int64
	cases := []struct {
		name string
		vals pairs
	}{
		{
			name: "empty",
		},
		{
			name: "one",
			vals: pairs{{1, 2}},
		},
		{
			name: "one_zero",
			vals: pairs{{0, 2}},
		},
		{
			name: "two",
			vals: pairs{{1, 2}, {3, 4}},
		},
		{
			name: "two_zero",
			vals: pairs{{1, 2}, {0, 4}},
		},
		{
			name: "ten",
			vals: pairs{{1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}, {6, 6}, {7, 7}, {8, 8}, {9, 9}, {10, 10}},
		},
		{
			name: "ten_zero",
			vals: pairs{{1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}, {6, 6}, {7, 7}, {8, 8}, {9, 9}, {10, 10}, {0, 11}},
		},
	}

	runTest := func(t *testing.T, m *Map[int64, int64], vals pairs) {
		for i, pair := range vals {
			m.Put(pair[0], pair[1])
			if sz := m.Len(); sz != i+1 {
				t.Fatalf("unexpected size after %d put()s: %d", sz, i+1)
			}
		}
		for i, pair := range vals {
			val, ok := m.Get(pair[0])
			if !ok {
				t.Fatalf("key number %d not found: %d", i, pair[0])
			}
			if val != pair[1] {
				t.Fatalf("incorrect value %d for key %d, expected %d", pair[1], pair[0], val)
			}
		}
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("zero_cap", func(t *testing.T) {
				m := New[int64, int64](0)
				runTest(t, m, tc.vals)
			})
			t.Run("full_cap", func(t *testing.T) {
				m := New[int64, int64](len(tc.vals))
				runTest(t, m, tc.vals)
			})
		})
	}
}

func TestMap64Delete(t *testing.T) {
	m := New[int, int](10)
	for i := 0; i < 100; i++ {
		m.Put(i, -i)
	}
	if sz := m.Len(); sz != 100 {
		t.Fatalf("expected %d elements in map: %d", 100, sz)
	}

	for i := 0; i < 100; i++ {
		if found := m.Del(i); !found {
			t.Fatalf("deleted key should have been there: %d", i)
		}
		if sz := m.Len(); sz != 100-i-1 {
			t.Fatalf("expected %d elements in map: %d", 100-i-1, sz)
		}
		if found := m.Del(i); found {
			t.Fatalf("deleted key should not be there: %d", i)
		}
	}

	if sz := m.Len(); sz != 0 {
		t.Fatalf("map not empty, %d elements remain", sz)
	}
}
