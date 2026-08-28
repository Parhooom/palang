package parser

type Type uint8

const (
	_ Type = iota
	TYPE_VOID
	TYPE_INT
)

var offsetOf = map[Type]uint{
	TYPE_VOID: 0,
	TYPE_INT:  8,
}
