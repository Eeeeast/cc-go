package lex

import (
	"sort"
)

const daRoot = 1

type DoubleArrayTrie struct {
	base   []int
	check  []int
	output []TokenType
	piece  []PuzzlePiece
}

type trieNode struct {
	children map[int]*trieNode
	ttype    TokenType
	piece    PuzzlePiece
}

var defaultTrie = buildLanguageTrie()

func buildLanguageTrie() *DoubleArrayTrie {
	entries := []struct {
		text  string
		ttype TokenType
		piece PuzzlePiece
	}{
		// Действия и декларации (>>)
		{"for", TokenKwFor, PieceAction},
		{"if", TokenKwIf, PieceAction},
		{"else", TokenKwElse, PieceAction},
		{"func", TokenKwFunc, PieceAction},
		{"struct", TokenKwStruct, PieceAction},
		{"return", TokenKwReturn, PieceAction},
		{"type", TokenKwType, PieceAction},
		{"var", TokenKwVar, PieceAction},

		// Терминаторы (>|) и булевы константы (><)
		{"break", TokenKwBreak, PieceTerminator},
		{"continue", TokenKwContinue, PieceTerminator},
		{"true", TokenBool, PieceOperand},
		{"false", TokenBool, PieceOperand},

		// Операторы и связки (<> и >>)
		{"+", TokenPlus, PieceConnector},
		{"-", TokenMinus, PieceConnector},
		{"*", TokenStar, PieceConnector},
		{"/", TokenSlash, PieceConnector},
		{"%", TokenPercent, PieceConnector},
		{"=", TokenAssign, PieceConnector},
		{":=", TokenColonEq, PieceConnector},
		{"==", TokenEq, PieceConnector},
		{"!=", TokenNotEq, PieceConnector},
		{"<", TokenLt, PieceConnector},
		{"<=", TokenLtEq, PieceConnector},
		{">", TokenGt, PieceConnector},
		{">=", TokenGtEq, PieceConnector},
		{"&&", TokenAnd, PieceConnector},
		{"||", TokenOr, PieceConnector},
		{"!", TokenNot, PieceAction},
		{".", TokenDot, PieceConnector},
		{",", TokenComma, PieceConnector},

		// Разделители (>|)
		{";", TokenSemicolon, PieceTerminator},
		{"(", TokenLParen, PieceTerminator},
		{")", TokenRParen, PieceTerminator},
		{"{", TokenLBrace, PieceTerminator},
		{"}", TokenRBrace, PieceTerminator},
		{"[", TokenLBracket, PieceTerminator},
		{"]", TokenRBracket, PieceTerminator},
	}

	root := &trieNode{children: make(map[int]*trieNode)}
	for _, e := range entries {
		curr := root
		for i := 0; i < len(e.text); i++ {
			c := int(e.text[i])
			if curr.children[c] == nil {
				curr.children[c] = &trieNode{children: make(map[int]*trieNode)}
			}
			curr = curr.children[c]
		}
		curr.ttype, curr.piece = e.ttype, e.piece
	}

	da := &DoubleArrayTrie{
		base:   make([]int, 154),
		check:  make([]int, 154),
		output: make([]TokenType, 154),
		piece:  make([]PuzzlePiece, 154),
	}
	da.check[daRoot] = -1

	var build func(state int, node *trieNode)
	build = func(state int, node *trieNode) {
		if len(node.children) == 0 {
			return
		}
		keys := make([]int, 0, len(node.children))
		for k := range node.children {
			keys = append(keys, k)
		}
		sort.Ints(keys)

		b := 2
		for {
			ok := true
			for _, k := range keys {
				if da.check[b+k] != 0 {
					ok = false
					break
				}
			}
			if ok {
				break
			}
			b++
		}

		da.base[state] = b
		for _, k := range keys {
			next, child := b+k, node.children[k]
			da.check[next], da.output[next], da.piece[next] = state, child.ttype, child.piece
		}
		for _, k := range keys {
			build(b+k, node.children[k])
		}
	}
	build(daRoot, root)

	return da
}

func (d *DoubleArrayTrie) Step(state int, ch rune) (int, bool) {
	if ch <= 0 || ch >= 128 || state <= 0 || state >= len(d.base) || d.base[state] == 0 {
		return 0, false
	}
	next := d.base[state] + int(ch)
	if next > 0 && next < len(d.check) && d.check[next] == state {
		return next, true
	}
	return 0, false
}

func IsReserved(ident string) bool {
	state := daRoot
	for _, ch := range ident {
		next, ok := defaultTrie.Step(state, ch)
		if !ok {
			return false
		}
		state = next
	}
	tt := defaultTrie.output[state]
	return tt >= TokenKwBreak && tt <= TokenKwVar
}
