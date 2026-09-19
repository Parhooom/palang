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

// type Expr interface {
// 	Type() Type
// }
//
// type Int struct {
// 	value int64
// }
//
// func (i Int) Type() Type {
// 	return TYPE_INT
// }
//
// type Bool struct {
// 	value bool
// }
//
// func (b Bool) Type() Type {
// 	return TYPE_BOOL
// }
//
// type BinaryExpr struct {
// 	lhs Expr
// 	rhs Expr
// 	op  string
// }
