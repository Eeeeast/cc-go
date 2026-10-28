package lex

type TokenType int

const (
	TokenNone TokenType = iota
	TokenEOF

	// Динамические токены
	TokenIdent
	TokenInt
	TokenString
	TokenBool

	// Пробельные токены
	TokenCommentLine
	TokenCommentBlock

	// Ключевые слова
	TokenKwBreak
	TokenKwContinue
	TokenKwFor
	TokenKwIf
	TokenKwElse
	TokenKwFunc
	TokenKwStruct
	TokenKwReturn
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
