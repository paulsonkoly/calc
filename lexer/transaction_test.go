package lexer_test

import (
	"testing"

	"github.com/phaul/calc/lexer"
	"github.com/stretchr/testify/assert"
)

func TestTransaction(t *testing.T) {
  input := "a = 2-3*(a+1)-2"
  tl := lexer.NewTLexer(input)

  tl.Snapshot()
  assert.True(t, tl.Next()) 
  assert.Equal(t, tl.Token().Value, "a")
  assert.True(t, tl.Next()) 
  assert.Equal(t, tl.Token().Value, "=")
  assert.True(t, tl.Next()) 
  assert.Equal(t, tl.Token().Value, "2")

  tl.Rollback()
  assert.True(t, tl.Next()) 
  assert.Equal(t, tl.Token().Value, "a")

  tl.Snapshot()
  assert.True(t, tl.Next()) 
  assert.Equal(t, tl.Token().Value, "=")
  assert.True(t, tl.Next()) 
  assert.Equal(t, tl.Token().Value, "2")

  tl.Snapshot()
  assert.True(t, tl.Next()) 
  assert.Equal(t, tl.Token().Value, "-")

  tl.Rollback()
  assert.True(t, tl.Next()) 
  assert.Equal(t, tl.Token().Value, "-")
}
