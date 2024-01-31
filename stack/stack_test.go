package stack_test

import (
	"testing"

	"github.com/phaul/calc/stack"
	"github.com/phaul/calc/types/value"
	"github.com/stretchr/testify/assert"
)

func TestStack(t *testing.T) {
	s := stack.NewStack()

	v, ok := s.LookUp("something")
	assert.Equal(t, value.Error("something not defined"), v)
	assert.False(t, ok)

	s.Set("a", value.Int(1))
	s.Set("b", value.Int(2))
	s.Set("c", value.Int(3))

	v, ok = s.LookUp("b")
	assert.Equal(t, value.Int(2), v)
	assert.True(t, ok)

	s.Push(nil)

	v, ok = s.LookUp("a")
	assert.Equal(t, value.Int(1), v)
	assert.True(t, ok)

	s.Set("d", value.Int(4))

	v, ok = s.LookUp("d")
	assert.Equal(t, value.Int(4), v)
	assert.True(t, ok)

	s.Set("c", value.Int(5))

	s.Pop()

	v, ok = s.LookUp("d")
	assert.Equal(t, value.Error("d not defined"), v)
	assert.False(t, ok)

	v, ok = s.LookUp("c")
	assert.Equal(t, value.Int(3), v)
	assert.True(t, ok)
}