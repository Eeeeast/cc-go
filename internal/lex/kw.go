package lex

import (
	"maps"
	"slices"
)

const daRoot = 1
const daCapacity = 256

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

type trieEntry struct {
	text  string
	ttype TokenType
}

func buildPrefixTree(entries []trieEntry) *trieNode {
	root := &trieNode{children: make(map[int]*trieNode)}
	for _, e := range entries {
		curr := root
		for i := range len(e.text) {
			c := int(e.text[i])
			if curr.children[c] == nil {
				curr.children[c] = &trieNode{children: make(map[int]*trieNode)}
			}
			curr = curr.children[c]
		}
		curr.ttype = e.ttype
	}
	return root
}

func findBase(check []int, keys []int) int {
	for b := 2; ; b++ {
		ok := true
		for _, k := range keys {
			if b+k >= len(check) {
				panic("lex: DoubleArrayTrie capacity exceeded")
			}
			if check[b+k] != 0 {
				ok = false
				break
			}
		}
		if ok {
			return b
		}
	}
}

func buildLanguageTrie() *DoubleArrayTrie {
	entries := []trieEntry{
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
		{"true", TokenKwTrue},
		{"false", TokenKwFalse},

		// Связки
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

	root := buildPrefixTree(entries)

	da := &DoubleArrayTrie{
		base:   make([]int, daCapacity),
		check:  make([]int, daCapacity),
		output: make([]TokenType, daCapacity),
	}
	da.check[daRoot] = -1

	var build func(state int, node *trieNode)
	build = func(state int, node *trieNode) {
		if len(node.children) == 0 {
			return
		}
		keys := slices.Sorted(maps.Keys(node.children))
		b := findBase(da.check, keys)

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
