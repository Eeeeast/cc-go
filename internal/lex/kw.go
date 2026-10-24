package lex

import (
	"sort"
)

const daRoot = 1

type DoubleArrayTrie struct {
	base   []int
	check  []int
	output []TokenType
}

type trieNode struct {
	children map[int]*trieNode
	ttype    TokenType
}

var defaultTrie = buildLanguageTrie()

func buildLanguageTrie() *DoubleArrayTrie {
	entries := []struct {
		text  string
		ttype TokenType
	}{
		// Действия и декларации
		{"for", TokenKwFor},
		{"if", TokenKwIf},
		{"else", TokenKwElse},
		{"func", TokenKwFunc},
		{"struct", TokenKwStruct},
		{"return", TokenKwReturn},
		{"type", TokenKwType},
		{"var", TokenKwVar},

		// Оператор
		{"!", TokenNot},

		// Булевы константы
		{"true", TokenBool},
		{"false", TokenBool},

		// Cвязки
		{"+", TokenPlus},
		{"-", TokenMinus},
		{"*", TokenStar},
		{"/", TokenSlash},
		{"%", TokenPercent},
		{"=", TokenAssign},
		{":=", TokenColonEq},
		{"==", TokenEq},
		{"!=", TokenNotEq},
		{"<", TokenLt},
		{"<=", TokenLtEq},
		{">", TokenGt},
		{">=", TokenGtEq},
		{"&&", TokenAnd},
		{"||", TokenOr},
		{".", TokenDot},
		{",", TokenComma},

		// Терминаторы и разделители
		{"break", TokenKwBreak},
		{"continue", TokenKwContinue},
		{";", TokenSemicolon},
		{"(", TokenLParen},
		{")", TokenRParen},
		{"{", TokenLBrace},
		{"}", TokenRBrace},
		{"[", TokenLBracket},
		{"]", TokenRBracket},
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
		curr.ttype = e.ttype
	}

	da := &DoubleArrayTrie{
		base:   make([]int, 154),
		check:  make([]int, 154),
		output: make([]TokenType, 154),
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
			da.check[next], da.output[next] = state, child.ttype
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
