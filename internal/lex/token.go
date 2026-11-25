package lex

import "fmt"

type TokenType int

const (
	TokenNone TokenType = iota
	TokenEOF

	// Динамические токены

	TokenIdent
	TokenInt
	TokenString

	// Ключевые слова

	TokenKwBreak
	TokenKwContinue
	TokenKwFor
	TokenKwIf
	TokenKwElse
	TokenKwFunc
	TokenKwStruct
	TokenKwReturn
	TokenKwTrue
	TokenKwFalse
	TokenKwType
	TokenKwVar

	// Операторы и логические отношения

	TokenPlus    // +
	TokenMinus   // -
	TokenStar    // *
	TokenSlash   // /
	TokenPercent // %
	TokenAssign  // =
	TokenColonEq // :=
	TokenEq      // ==
	TokenNotEq   // !=
	TokenLt      // <
	TokenLtEq    // <=
	TokenGt      // >
	TokenGtEq    // >=
	TokenAnd     // &&
	TokenOr      // ||
	TokenNot     // !
	TokenComma   // ,
	TokenDot     // .

	// Разделители и терминаторы

	TokenLParen    // (
	TokenRParen    // )
	TokenLBrace    // {
	TokenRBrace    // }
	TokenLBracket  // [
	TokenRBracket  // ]
	TokenSemicolon // ;
)

type Token struct {
	Type   TokenType
	Lexeme string
	Pos    SrcPos
}

// `LineNo` и `ColNo` 1-based, 0 - нет данных
type SrcPos struct {
	SrcName string
	LineNo  int
	ColNo   int
}

func NewSrcPos(srcName string, lineNo int, colNo int) SrcPos {
	return SrcPos{
		SrcName: srcName,
		LineNo:  lineNo,
		ColNo:   colNo,
	}
}

func (p SrcPos) FormatError(msg string) string {
	if p.SrcName == "" {
		return msg
	}
	if p.LineNo == 0 {
		return fmt.Sprintf("%s: %s", p.SrcName, msg)
	}
	if p.ColNo == 0 {
		return fmt.Sprintf("%s:%d: %s", p.SrcName, p.LineNo, msg)
	}
	return fmt.Sprintf("%s:%d:%d: %s", p.SrcName, p.LineNo, p.ColNo, msg)
}
