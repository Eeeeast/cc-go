package lexer

import (
	"errors"
	"fmt"
)

var ErrTypeMismatch = errors.New("token type mismatch")

type Symbol uint64

const (
	SymError Symbol = iota

	// Delimiters
	SymOpenParen    // (
	SymCloseParen   // )
	SymOpenBrace    // {
	SymCloseBrace   // }
	SymOpenBracket  // [
	SymCloseBracket // ]

	// Punctuation
	SymEq         // =
	SymEqEq       // ==
	SymNotEq      // !=
	SymLt         // <
	SymLtEq       // <=
	SymGt         // >
	SymGtEq       // >=
	SymPlus       // +
	SymMinus      // -
	SymStar       // *
	SymSlash      // /
	SymPercent    // %
	SymAndAnd     // &&
	SymOrOr       // ||
	SymNot        // !
	SymArrow      // ->
	SymSemi       // ;
	SymComma      // ,
	SymColon      // :
	SymColonColon // ::
	SymDot        // .

	// Keywords
	SymUnderscore // _
	SymAs         // as
	SymAsync      // async
	SymAwait      // await
	SymBreak      // break
	SymConst      // const
	SymContinue   // continue
	SymCrate      // crate
	SymDyn        // dyn
	SymElse       // else
	SymEnum       // enum
	SymExtern     // extern
	SymFalse      // false
	SymFn         // fn
	SymFor        // for
	SymIf         // if
	SymImpl       // impl
	SymIn         // in
	SymLet        // let
	SymLoop       // loop
	SymMatch      // match
	SymMod        // mod
	SymMove       // move
	SymMut        // mut
	SymPub        // pub
	SymRef        // ref
	SymReturn     // return
	SymSelfLower  // self
	SymSelfUpper  // Self
	SymStatic     // static
	SymStruct     // struct
	SymSuper      // super
	SymTrait      // trait
	SymTrue       // true
	TokenType     // type
	SymUnsafe     // unsafe
	SymUse        // use
	SymWhere      // where
	SymWhile      // while
	SymAbstract   // abstract
	SymBecome     // become
	SymBox        // box
	SymDo         // do
	SymFinal      // final
	SymGen        // gen
	SymMacro      // macro
	SymOverride   // override
	SymPriv       // priv
	SymTry        // try
	SymTypeof     // typeof
	SymUnsized    // unsized
	SymVirtual    // virtual
	SymYield      // yield

	firstDynamicSymbol
)

type Interner struct {
	strings []string          // ID -> string
	lookup  map[string]Symbol // string -> ID
}

func NewInterner() *Interner {
	in := &Interner{
		strings: make([]string, 0, 128),
		lookup:  make(map[string]Symbol, 128),
	}

	reg := func(sym Symbol, s string) {
		for len(in.strings) <= int(sym) {
			in.strings = append(in.strings, "")
		}
		in.strings[sym] = s
		in.lookup[s] = sym
	}

	reg(SymOpenParen, "(")
	reg(SymCloseParen, ")")
	reg(SymOpenBrace, "{")
	reg(SymCloseBrace, "}")
	reg(SymOpenBracket, "[")
	reg(SymCloseBracket, "]")

	reg(SymEq, "=")
	reg(SymEqEq, "==")
	reg(SymNotEq, "!=")
	reg(SymLt, "<")
	reg(SymLtEq, "<=")
	reg(SymGt, ">")
	reg(SymGtEq, ">=")
	reg(SymPlus, "+")
	reg(SymMinus, "-")
	reg(SymStar, "*")
	reg(SymSlash, "/")
	reg(SymPercent, "%")
	reg(SymAndAnd, "&&")
	reg(SymOrOr, "||")
	reg(SymNot, "!")
	reg(SymArrow, "->")
	reg(SymSemi, ";")
	reg(SymComma, ",")
	reg(SymColon, ":")
	reg(SymColonColon, "::")
	reg(SymDot, ".")

	reg(SymUnderscore, "_")
	reg(SymAs, "as")
	reg(SymAsync, "async")
	reg(SymAwait, "await")
	reg(SymBreak, "break")
	reg(SymConst, "const")
	reg(SymContinue, "continue")
	reg(SymCrate, "crate")
	reg(SymDyn, "dyn")
	reg(SymElse, "else")
	reg(SymEnum, "enum")
	reg(SymExtern, "extern")
	reg(SymFalse, "false")
	reg(SymFn, "fn")
	reg(SymFor, "for")
	reg(SymIf, "if")
	reg(SymImpl, "impl")
	reg(SymIn, "in")
	reg(SymLet, "let")
	reg(SymLoop, "loop")
	reg(SymMatch, "match")
	reg(SymMod, "mod")
	reg(SymMove, "move")
	reg(SymMut, "mut")
	reg(SymPub, "pub")
	reg(SymRef, "ref")
	reg(SymReturn, "return")
	reg(SymSelfLower, "self")
	reg(SymSelfUpper, "Self")
	reg(SymStatic, "static")
	reg(SymStruct, "struct")
	reg(SymSuper, "super")
	reg(SymTrait, "trait")
	reg(SymTrue, "true")
	reg(TokenType, "type")
	reg(SymUnsafe, "unsafe")
	reg(SymUse, "use")
	reg(SymWhere, "where")
	reg(SymWhile, "while")
	reg(SymAbstract, "abstract")
	reg(SymBecome, "become")
	reg(SymBox, "box")
	reg(SymDo, "do")
	reg(SymFinal, "final")
	reg(SymGen, "gen")
	reg(SymMacro, "macro")
	reg(SymOverride, "override")
	reg(SymPriv, "priv")
	reg(SymTry, "try")
	reg(SymTypeof, "typeof")
	reg(SymUnsized, "unsized")
	reg(SymVirtual, "virtual")
	reg(SymYield, "yield")

	return in
}

func (in *Interner) Intern(s string) Symbol {
	if id, ok := in.lookup[s]; ok {
		return id
	}
	id := Symbol(len(in.strings))
	in.strings = append(in.strings, s)
	in.lookup[s] = id
	return id
}

func (in *Interner) Resolve(sym Symbol) string {
	if int(sym) < len(in.strings) {
		return in.strings[sym]
	}
	return fmt.Sprintf("<invalid:%d>", sym)
}
