package combinator

// Error indicates an error in the input source code
type Error struct {
	from    int
	to      int
	message string
}

func (e *Error) Error() string { return e.message }

// From points to the starting position of the token in the buffer
func (e *Error) From() int { return e.from }

// To points to the ending position of the token in the buffer
func (e *Error) To() int { return e.to }

// Message describes the error
func (e *Error) Message() string { return e.message }
