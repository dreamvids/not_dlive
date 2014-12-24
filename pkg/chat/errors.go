package chat

import "errors"

var (
	InvalidMessageError = errors.New("Invalid message")
)

// Database error
type DatabaseError struct {
	err string
}

// Error string
func (this *DatabaseError) Error() string {
	return this.err
}

// Protocol error
type ProtocolError struct {
	err string
}

// Error string
func (this *ProtocolError) Error() string {
	return this.err
}
