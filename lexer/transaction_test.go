package lexer_test

import (
	"testing"

	"github.com/paulsonkoly/calc/lexer"
	"github.com/paulsonkoly/calc/types/token"
	"github.com/stretchr/testify/assert"
)

func TestTransaction(t *testing.T) {
	input := "a = 2-3*(a+1)-2"
	tl := lexer.NewTLexer(input)

	tl.Snapshot()
	assert.True(t, tl.Next())
	assert.Equal(t, "a", tl.Token().(token.Type).Value)
	assert.True(t, tl.Next())
	assert.Equal(t, "=", tl.Token().(token.Type).Value)
	tl.Snapshot()
	assert.True(t, tl.Next())
	assert.Equal(t, "2", tl.Token().(token.Type).Value)

	tl.Rollback()
	tl.Rollback()
	assert.True(t, tl.Next())
	assert.Equal(t, "a", tl.Token().(token.Type).Value)

	tl.Snapshot()
	assert.True(t, tl.Next())
	assert.Equal(t, "=", tl.Token().(token.Type).Value)
	assert.True(t, tl.Next())
	assert.Equal(t, "2", tl.Token().(token.Type).Value)

	tl.Snapshot()
	assert.True(t, tl.Next())
	assert.Equal(t, "-", tl.Token().(token.Type).Value)

	tl.Rollback()
	assert.True(t, tl.Next())
	assert.Equal(t, "-", tl.Token().(token.Type).Value)
}
