package ecs

import "testing"

func TestBitwiseClear(t *testing.T) {
	type testStruct struct {
		input    archetypeMask
		clear    archetypeMask
		expected archetypeMask
	}

	tests := []testStruct{
		{
			input:    buildArchMaskFromId(0, 1, 2, 3, 4),
			clear:    buildArchMaskFromId(2, 3, 4),
			expected: buildArchMaskFromId(0, 1),
		},
		{
			input:    buildArchMaskFromId(0, 1),
			clear:    buildArchMaskFromId(1),
			expected: buildArchMaskFromId(0),
		},
		{
			input:    buildArchMaskFromId(0, 1),
			clear:    buildArchMaskFromId(0, 1),
			expected: buildArchMaskFromId(),
		},
		{
			input:    buildArchMaskFromId(0, 1),
			clear:    buildArchMaskFromId(0, 1, 2),
			expected: buildArchMaskFromId(),
		},
		{
			input:    buildArchMaskFromId(0, 1),
			clear:    buildArchMaskFromId(2),
			expected: buildArchMaskFromId(0, 1),
		},
	}

	for _, test := range tests {
		got := test.input.bitwiseClear(test.clear)
		if got != test.expected {
			t.Errorf("error")
		}
	}
}

func TestBitwiseOr(t *testing.T) {
	type testStruct struct {
		input    archetypeMask
		clear    archetypeMask
		expected archetypeMask
	}

	tests := []testStruct{
		{
			input:    buildArchMaskFromId(0, 1, 2, 3, 4),
			clear:    buildArchMaskFromId(2, 3, 4),
			expected: buildArchMaskFromId(0, 1, 2, 3, 4),
		},
		{
			input:    buildArchMaskFromId(0, 1),
			clear:    buildArchMaskFromId(1),
			expected: buildArchMaskFromId(0, 1),
		},
		{
			input:    buildArchMaskFromId(0, 1),
			clear:    buildArchMaskFromId(0, 1),
			expected: buildArchMaskFromId(0, 1),
		},
		{
			input:    buildArchMaskFromId(0, 1),
			clear:    buildArchMaskFromId(0, 1, 2),
			expected: buildArchMaskFromId(0, 1, 2),
		},
		{
			input:    buildArchMaskFromId(0, 1),
			clear:    buildArchMaskFromId(2),
			expected: buildArchMaskFromId(0, 1, 2),
		},
		{
			input:    buildArchMaskFromId(0, 1),
			clear:    buildArchMaskFromId(2, 3),
			expected: buildArchMaskFromId(0, 1, 2, 3),
		},
		{
			input:    buildArchMaskFromId(),
			clear:    buildArchMaskFromId(2, 3),
			expected: buildArchMaskFromId(2, 3),
		},
		{
			input:    buildArchMaskFromId(2, 3),
			clear:    buildArchMaskFromId(),
			expected: buildArchMaskFromId(2, 3),
		},
		{
			input:    buildArchMaskFromId(),
			clear:    buildArchMaskFromId(),
			expected: buildArchMaskFromId(),
		},
	}

	for _, test := range tests {
		got := test.input.bitwiseOr(test.clear)
		if got != test.expected {
			t.Errorf("error")
		}
	}
}

func TestBitwiseAnd(t *testing.T) {
	type testStruct struct {
		input    archetypeMask
		clear    archetypeMask
		expected archetypeMask
	}

	tests := []testStruct{
		{
			input:    buildArchMaskFromId(0, 1, 2, 3, 4),
			clear:    buildArchMaskFromId(2, 3, 4),
			expected: buildArchMaskFromId(2, 3, 4),
		},
		{
			input:    buildArchMaskFromId(0, 1),
			clear:    buildArchMaskFromId(1),
			expected: buildArchMaskFromId(1),
		},
		{
			input:    buildArchMaskFromId(0, 1),
			clear:    buildArchMaskFromId(0, 1),
			expected: buildArchMaskFromId(0, 1),
		},
		{
			input:    buildArchMaskFromId(0, 1),
			clear:    buildArchMaskFromId(0, 1, 2),
			expected: buildArchMaskFromId(0, 1),
		},
		{
			input:    buildArchMaskFromId(0, 1),
			clear:    buildArchMaskFromId(2),
			expected: buildArchMaskFromId(),
		},
		{
			input:    buildArchMaskFromId(0, 1),
			clear:    buildArchMaskFromId(2, 3),
			expected: buildArchMaskFromId(),
		},
		{
			input:    buildArchMaskFromId(),
			clear:    buildArchMaskFromId(2, 3),
			expected: buildArchMaskFromId(),
		},
		{
			input:    buildArchMaskFromId(2, 3),
			clear:    buildArchMaskFromId(),
			expected: buildArchMaskFromId(),
		},
		{
			input:    buildArchMaskFromId(),
			clear:    buildArchMaskFromId(),
			expected: buildArchMaskFromId(),
		},
		{
			input:    buildArchMaskFromId(0, 2, 4),
			clear:    buildArchMaskFromId(1, 2, 3, 4),
			expected: buildArchMaskFromId(2, 4),
		},
	}

	for _, test := range tests {
		got := test.input.bitwiseAnd(test.clear)
		if got != test.expected {
			t.Errorf("error")
		}
	}
}

func TestAddCompMask(t *testing.T) {
	type testStruct struct {
		input    archetypeMask
		add CompId
		expected archetypeMask
	}

	tests := []testStruct{
		{
			input:    buildArchMaskFromId(0),
			add: 1,
			expected: buildArchMaskFromId(0, 1),
		},
		{
			input:    buildArchMaskFromId(0, 1),
			add: 3,
			expected: buildArchMaskFromId(0, 1, 3),
		},
		{
			input:    buildArchMaskFromId(0, 1),
			add: 63,
			expected: buildArchMaskFromId(0, 1, 63),
		},
		{
			input:    buildArchMaskFromId(0, 1),
			add: 1,
			expected: buildArchMaskFromId(0, 1),
		},
	}

	for _, test := range tests {
		test.input.addComponent(test.add)
		got := test.input
		if got != test.expected {
			t.Errorf("error")
		}
	}
}

func TestRemoveCompMask(t *testing.T) {
	type testStruct struct {
		input    archetypeMask
		remove CompId
		expected archetypeMask
	}

	tests := []testStruct{
		{
			input:    buildArchMaskFromId(0),
			remove: 1,
			expected: buildArchMaskFromId(0),
		},
		{
			input:    buildArchMaskFromId(0, 1),
			remove: 63,
			expected: buildArchMaskFromId(0, 1),
		},
		{
			input:    buildArchMaskFromId(0),
			remove: 0,
			expected: buildArchMaskFromId(),
		},
		{
			input:    buildArchMaskFromId(0, 1),
			remove: 1,
			expected: buildArchMaskFromId(0),
		},
		{
			input:    buildArchMaskFromId(0, 1),
			remove: 0,
			expected: buildArchMaskFromId(1),
		},
	}

	for _, test := range tests {
		test.input.removeComponent(test.remove)
		got := test.input
		if got != test.expected {
			t.Errorf("error")
		}
	}
}
