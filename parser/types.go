package parser

type Type uint8

const (
	_ Type = iota
	TYPE_VOID
	TYPE_INT
	TYPE_BOOL
)

var offsetOf = map[Type]uint{
	TYPE_VOID: 0,
	TYPE_INT:  8,
	TYPE_BOOL: 8,
}
