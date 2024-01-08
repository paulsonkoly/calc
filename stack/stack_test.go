package stack_test

import (
	"testing"

	"github.com/phaul/calc/types"
	"github.com/phaul/calc/stack"
	"github.com/stretchr/testify/assert"
)

func TestStack(t * testing.T) {
	s := stack.NewStack()

	v, ok := s.LookUp("something")
	assert.Equal(t, types.ValueError("something not defined"), v)
	assert.False(t, ok)

	s.Set("a", types.ValueInt(1))
	s.Set("b", types.ValueInt(2))
	s.Set("c", types.ValueInt(3))

	v, ok = s.LookUp("b")
	assert.Equal(t, types.ValueInt(2), v)
	assert.True(t, ok)

	s.Push()

	v, ok = s.LookUp("a")
	assert.Equal(t, types.ValueInt(1), v)
	assert.True(t, ok)

	s.Set("d", types.ValueInt(4))

	v, ok = s.LookUp("d")
	assert.Equal(t, types.ValueInt(4), v)
	assert.True(t, ok)

	s.Set("c", types.ValueInt(5))

	s.Pop()

	v, ok = s.LookUp("d")
	assert.Equal(t, types.ValueError("d not defined"), v)
	assert.False(t, ok)

	v, ok = s.LookUp("c")
	assert.Equal(t, types.ValueInt(3), v)
	assert.True(t, ok)
}
