package ast

import "compiler/internal/lex"

type Node interface {
	Pos() lex.SrcPos
	Accept(v Visitor) error
}

type Expr interface {
	Node
	exprNode()
}

type Stmt interface {
	Node
	stmtNode()
}

type Decl interface {
	Node
	declNode()
}

type Visitor interface {
	VisitProgram(n *Program) error
	VisitVarDecl(n *VarDecl) error
	VisitFuncDecl(n *FuncDecl) error
	VisitBlockStmt(n *BlockStmt) error
	VisitAssignStmt(n *AssignStmt) error
	VisitReturnStmt(n *ReturnStmt) error
	VisitExprStmt(n *ExprStmt) error
	VisitIntLiteral(n *IntLiteral) error
	VisitIdentExpr(n *IdentExpr) error
	VisitUnaryExpr(n *UnaryExpr) error
	VisitBinaryExpr(n *BinaryExpr) error
	VisitCallExpr(n *CallExpr) error
}

type Program struct {
	Decls  []Decl
	SrcPos lex.SrcPos
}

func (n *Program) Pos() lex.SrcPos        { return n.SrcPos }
func (n *Program) Accept(v Visitor) error { return v.VisitProgram(n) }

type VarDecl struct {
	Name   string
	Type   string
	Init   Expr
	SrcPos lex.SrcPos
}

func (n *VarDecl) Pos() lex.SrcPos        { return n.SrcPos }
func (n *VarDecl) Accept(v Visitor) error { return v.VisitVarDecl(n) }
func (n *VarDecl) declNode()              {}
func (n *VarDecl) stmtNode()              {}

type Param struct {
	Name   string
	Type   string
	SrcPos lex.SrcPos
}

type FuncDecl struct {
	Name       string
	Params     []Param
	ReturnType string // "" для функции без возвращаемого значения, "int" для int
	Body       *BlockStmt
	SrcPos     lex.SrcPos
}

func (n *FuncDecl) Pos() lex.SrcPos        { return n.SrcPos }
func (n *FuncDecl) Accept(v Visitor) error { return v.VisitFuncDecl(n) }
func (n *FuncDecl) declNode()              {}

type BlockStmt struct {
	Stmts  []Stmt
	SrcPos lex.SrcPos
}

func (n *BlockStmt) Pos() lex.SrcPos        { return n.SrcPos }
func (n *BlockStmt) Accept(v Visitor) error { return v.VisitBlockStmt(n) }
func (n *BlockStmt) stmtNode()              {}

type AssignStmt struct {
	Name   string
	Value  Expr
	SrcPos lex.SrcPos
}

func (n *AssignStmt) Pos() lex.SrcPos        { return n.SrcPos }
func (n *AssignStmt) Accept(v Visitor) error { return v.VisitAssignStmt(n) }
func (n *AssignStmt) stmtNode()              {}

type ReturnStmt struct {
	Value  Expr // nil для пустого `return;`
	SrcPos lex.SrcPos
}

func (n *ReturnStmt) Pos() lex.SrcPos        { return n.SrcPos }
func (n *ReturnStmt) Accept(v Visitor) error { return v.VisitReturnStmt(n) }
func (n *ReturnStmt) stmtNode()              {}

type ExprStmt struct {
	Expr    Expr
	SrcPos  lex.SrcPos
	SemiPos lex.SrcPos
}

func (n *ExprStmt) Pos() lex.SrcPos        { return n.SrcPos }
func (n *ExprStmt) Accept(v Visitor) error { return v.VisitExprStmt(n) }
func (n *ExprStmt) stmtNode()              {}

type IntLiteral struct {
	Value  int64
	SrcPos lex.SrcPos
}

func (n *IntLiteral) Pos() lex.SrcPos        { return n.SrcPos }
func (n *IntLiteral) Accept(v Visitor) error { return v.VisitIntLiteral(n) }
func (n *IntLiteral) exprNode()              {}

type IdentExpr struct {
	Name   string
	SrcPos lex.SrcPos
}

func (n *IdentExpr) Pos() lex.SrcPos        { return n.SrcPos }
func (n *IdentExpr) Accept(v Visitor) error { return v.VisitIdentExpr(n) }
func (n *IdentExpr) exprNode()              {}

type UnaryExpr struct {
	Op      lex.TokenType
	Operand Expr
	SrcPos  lex.SrcPos
}

func (n *UnaryExpr) Pos() lex.SrcPos        { return n.SrcPos }
func (n *UnaryExpr) Accept(v Visitor) error { return v.VisitUnaryExpr(n) }
func (n *UnaryExpr) exprNode()              {}

type BinaryExpr struct {
	Op     lex.TokenType
	Left   Expr
	Right  Expr
	SrcPos lex.SrcPos
}

func (n *BinaryExpr) Pos() lex.SrcPos        { return n.SrcPos }
func (n *BinaryExpr) Accept(v Visitor) error { return v.VisitBinaryExpr(n) }
func (n *BinaryExpr) exprNode()              {}

type CallExpr struct {
	Callee string
	Args   []Expr
	SrcPos lex.SrcPos
}

func (n *CallExpr) Pos() lex.SrcPos        { return n.SrcPos }
func (n *CallExpr) Accept(v Visitor) error { return v.VisitCallExpr(n) }
func (n *CallExpr) exprNode()              {}

var (
	_ Node = (*Program)(nil)
	_ Decl = (*VarDecl)(nil)
	_ Stmt = (*VarDecl)(nil)
	_ Decl = (*FuncDecl)(nil)
	_ Stmt = (*BlockStmt)(nil)
	_ Stmt = (*AssignStmt)(nil)
	_ Stmt = (*ReturnStmt)(nil)
	_ Stmt = (*ExprStmt)(nil)
	_ Expr = (*IntLiteral)(nil)
	_ Expr = (*IdentExpr)(nil)
	_ Expr = (*UnaryExpr)(nil)
	_ Expr = (*BinaryExpr)(nil)
	_ Expr = (*CallExpr)(nil)
)
