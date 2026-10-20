package lex

// PuzzlePiece определяет пазловую модель токенизации (jigsaw puzzle):
// - "><" (PieceOperand):    идентификаторы и литералы (операнды)
// - ">>" (PieceAction):     ключовые глаголы, деклараци и унарные модификаторы
// - "<>" (PieceConnector):  бинарные операторы и логические отношения (частицы)
// - ">|" (PieceTerminator): разделители блоков/инструкций и конец потока
type PuzzlePiece string

const (
	PieceNone       PuzzlePiece = ""
	PieceOperand    PuzzlePiece = "><"
	PieceAction     PuzzlePiece = ">>"
	PieceConnector  PuzzlePiece = "<>"
	PieceTerminator PuzzlePiece = ">|"
)

type TokenType int

const (
	TokenNone TokenType = iota
	TokenEOF

	// Динамические токены (><)
	TokenIdent
	TokenInt
	TokenString
	TokenBool

	// Пробельные токены автомата
	TokenCommentLine
	TokenCommentBlock

	// Ключевые слова (>> и >|)
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

	// Операторы и логические отношения (<> и >>)
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

	// Разделители и терминаторы (>|)
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
	Piece  PuzzlePiece
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
