package codegen

import (
	"fmt"
	"slices"

	"compiler/internal/ast"
	"compiler/internal/lex"
	"compiler/internal/vm"
)

type CompileError struct {
	Pos lex.SrcPos
	Msg string
}

func (e *CompileError) Error() string {
	return e.Pos.FormatError(e.Msg)
}

func newCompileError(pos lex.SrcPos, format string, args ...any) error {
	return &CompileError{
		Pos: pos,
		Msg: fmt.Sprintf(format, args...),
	}
}

type valType uint8

const (
	typeVoid valType = iota
	typeInt
)

type builtinSpec struct {
	op       vm.OpCode
	minArity int
	maxArity int // -1 для вариативных функций
	retType  valType
	passArgc bool
}

var builtins = map[string]builtinSpec{
	"fmt.Println": {op: vm.OpCallPrintln, minArity: 1, maxArity: 1, retType: typeVoid},
	"fmt.ReadInt": {op: vm.OpCallReadInt, minArity: 0, maxArity: 0, retType: typeInt},
	"abs":         {op: vm.OpCallAbs, minArity: 1, maxArity: 1, retType: typeInt},
	"min":         {op: vm.OpCallMin, minArity: 1, maxArity: -1, retType: typeInt, passArgc: true},
	"max":         {op: vm.OpCallMax, minArity: 1, maxArity: -1, retType: typeInt, passArgc: true},
}

var reservedNames = map[string]bool{
	"main": true,
	"int":  true,
	"fmt":  true,
	"abs":  true,
	"min":  true,
	"max":  true,
}

type funcSymbol struct {
	index   int
	arity   int
	retType valType
}

type funcContext struct {
	decl      *ast.FuncDecl
	sym       funcSymbol
	scopes    []map[string]int
	nextLocal int
}

type VMCodegen struct {
	code     []vm.Instruction
	fnInfos  []vm.FuncInfo
	globals  map[string]int
	funcs    map[string]funcSymbol
	curFunc  *funcContext
	lastType valType
}

var _ ast.Visitor = (*VMCodegen)(nil)

func New() *VMCodegen {
	return &VMCodegen{}
}

func Generate(prog *ast.Program) (*vm.Program, error) {
	return New().Generate(prog)
}

func (g *VMCodegen) Generate(prog *ast.Program) (*vm.Program, error) {
	g.code = nil
	g.fnInfos = nil
	g.globals = make(map[string]int)
	g.funcs = make(map[string]funcSymbol)
	g.curFunc = nil

	if err := prog.Accept(g); err != nil {
		return nil, err
	}

	return &vm.Program{
		Code:       g.code,
		Functions:  g.fnInfos,
		NumGlobals: len(g.globals),
	}, nil
}

func (g *VMCodegen) emit(op vm.OpCode, arg int64, pos lex.SrcPos) {
	g.code = append(g.code, vm.Instruction{Op: op, Arg: arg, Pos: pos})
}

func (g *VMCodegen) VisitProgram(n *ast.Program) error {
	var globalVars []*ast.VarDecl
	var funcDecls []*ast.FuncDecl

	// Проход 1: регистрация всех глобальных переменных и сигнатур функций.
	for _, decl := range n.Decls {
		switch d := decl.(type) {
		case *ast.VarDecl:
			if reservedNames[d.Name] {
				return newCompileError(d.SrcPos, "cannot declare variable with reserved name %q", d.Name)
			}
			if _, exists := g.globals[d.Name]; exists {
				return newCompileError(d.SrcPos, "global variable %q already declared", d.Name)
			}
			if _, exists := g.funcs[d.Name]; exists {
				return newCompileError(d.SrcPos, "name %q conflicts with function", d.Name)
			}
			g.globals[d.Name] = len(g.globals)
			globalVars = append(globalVars, d)

		case *ast.FuncDecl:
			if d.Name != "main" && reservedNames[d.Name] {
				return newCompileError(d.SrcPos, "cannot declare function with reserved name %q", d.Name)
			}
			if _, exists := g.funcs[d.Name]; exists {
				return newCompileError(d.SrcPos, "function %q already declared", d.Name)
			}
			if _, exists := g.globals[d.Name]; exists {
				return newCompileError(d.SrcPos, "name %q conflicts with global variable", d.Name)
			}

			retType := typeVoid
			if d.ReturnType == "int" {
				retType = typeInt
			}
			if d.Name == "main" && (len(d.Params) != 0 || retType != typeVoid) {
				return newCompileError(d.SrcPos, "function 'main' must have no parameters and no return value")
			}

			idx := len(g.fnInfos)
			g.funcs[d.Name] = funcSymbol{
				index:   idx,
				arity:   len(d.Params),
				retType: retType,
			}
			g.fnInfos = append(g.fnInfos, vm.FuncInfo{
				Name:  d.Name,
				Arity: len(d.Params),
			})
			funcDecls = append(funcDecls, d)
		}
	}

	mainSym, hasMain := g.funcs["main"]
	if !hasMain {
		return newCompileError(n.SrcPos, "missing entry point function 'main'")
	}

	// Проход 2: инициализация глобальных переменных, вызов main и остановка ВМ.
	for _, gv := range globalVars {
		if err := gv.Accept(g); err != nil {
			return err
		}
	}
	g.emit(vm.OpCall, int64(mainSym.index), n.SrcPos)
	g.emit(vm.OpHalt, 0, n.SrcPos)

	// Компиляция тел всех пользовательских функций.
	for _, fn := range funcDecls {
		if err := fn.Accept(g); err != nil {
			return err
		}
	}

	return nil
}

func (g *VMCodegen) VisitVarDecl(n *ast.VarDecl) error {
	// Глобальная переменная: слот уже выделен в проходе 1.
	if g.curFunc == nil {
		if n.Init == nil {
			return nil // нулевое значение 0 уже записано при создании v.globals
		}
		if err := g.emitIntExpr(n.Init); err != nil {
			return err
		}
		g.emit(vm.OpStoreGlobal, int64(g.globals[n.Name]), n.SrcPos)
		return nil
	}

	// Локальная переменная.
	if reservedNames[n.Name] {
		return newCompileError(n.SrcPos, "cannot declare variable with reserved name %q", n.Name)
	}

	// Сначала вычисляем инициализатор (имя новой переменной ещё не в области видимости).
	if n.Init != nil {
		if err := g.emitIntExpr(n.Init); err != nil {
			return err
		}
	} else {
		g.emit(vm.OpPush, 0, n.SrcPos)
	}

	curScope := g.curFunc.scopes[len(g.curFunc.scopes)-1]
	if _, exists := curScope[n.Name]; exists {
		return newCompileError(n.SrcPos, "variable %q already declared in this scope", n.Name)
	}

	slot := g.curFunc.nextLocal
	g.curFunc.nextLocal++
	curScope[n.Name] = slot

	g.emit(vm.OpStoreLocal, int64(slot), n.SrcPos)
	return nil
}

func (g *VMCodegen) VisitFuncDecl(n *ast.FuncDecl) error {
	sym := g.funcs[n.Name]
	g.fnInfos[sym.index].Addr = len(g.code)

	// Параметры и корневой блок тела функции образуют одну область видимости.
	rootScope := make(map[string]int, len(n.Params))
	ctx := &funcContext{
		decl:   n,
		sym:    sym,
		scopes: []map[string]int{rootScope},
	}

	for _, param := range n.Params {
		if reservedNames[param.Name] {
			return newCompileError(param.SrcPos, "cannot declare parameter with reserved name %q", param.Name)
		}
		if _, exists := rootScope[param.Name]; exists {
			return newCompileError(param.SrcPos, "duplicate parameter %q", param.Name)
		}
		rootScope[param.Name] = ctx.nextLocal
		ctx.nextLocal++
	}

	g.curFunc = ctx
	defer func() { g.curFunc = nil }()

	// Обходим инструкции тела напрямую, не создавая лишнюю вложенную область для внешнего блока.
	for _, stmt := range n.Body.Stmts {
		if err := stmt.Accept(g); err != nil {
			return err
		}
	}

	if sym.retType == typeInt {
		if !blockAlwaysReturns(n.Body) {
			return newCompileError(n.SrcPos, "function %q must return a value", n.Name)
		}
	} else {
		g.emit(vm.OpReturn, 0, n.Body.SrcPos)
	}

	g.fnInfos[sym.index].NumLocals = ctx.nextLocal
	return nil
}

func (g *VMCodegen) VisitBlockStmt(n *ast.BlockStmt) error {
	g.curFunc.scopes = append(g.curFunc.scopes, make(map[string]int))
	defer func() {
		g.curFunc.scopes = g.curFunc.scopes[:len(g.curFunc.scopes)-1]
	}()

	for _, stmt := range n.Stmts {
		if err := stmt.Accept(g); err != nil {
			return err
		}
	}
	return nil
}

func (g *VMCodegen) VisitAssignStmt(n *ast.AssignStmt) error {
	slot, isLocal, ok := g.resolveVar(n.Name)
	if !ok {
		return newCompileError(n.SrcPos, "undeclared variable %q", n.Name)
	}
	if err := g.emitIntExpr(n.Value); err != nil {
		return err
	}
	if isLocal {
		g.emit(vm.OpStoreLocal, int64(slot), n.SrcPos)
	} else {
		g.emit(vm.OpStoreGlobal, int64(slot), n.SrcPos)
	}
	return nil
}

func (g *VMCodegen) VisitReturnStmt(n *ast.ReturnStmt) error {
	if g.curFunc.sym.retType == typeVoid {
		if n.Value != nil {
			return newCompileError(n.SrcPos, "function %q does not return a value", g.curFunc.decl.Name)
		}
		g.emit(vm.OpReturn, 0, n.SrcPos)
		return nil
	}

	if n.Value == nil {
		return newCompileError(n.SrcPos, "function %q must return an int value", g.curFunc.decl.Name)
	}
	if err := g.emitIntExpr(n.Value); err != nil {
		return err
	}
	g.emit(vm.OpReturnVal, 0, n.SrcPos)
	return nil
}

func (g *VMCodegen) VisitExprStmt(n *ast.ExprStmt) error {
	if err := n.Expr.Accept(g); err != nil {
		return err
	}
	if g.lastType == typeInt {
		g.emit(vm.OpPop, 0, n.SemiPos)
	}
	return nil
}

func (g *VMCodegen) VisitIntLiteral(n *ast.IntLiteral) error {
	g.emit(vm.OpPush, n.Value, n.SrcPos)
	g.lastType = typeInt
	return nil
}

func (g *VMCodegen) VisitIdentExpr(n *ast.IdentExpr) error {
	slot, isLocal, ok := g.resolveVar(n.Name)
	if !ok {
		return newCompileError(n.SrcPos, "undeclared variable %q", n.Name)
	}
	if isLocal {
		g.emit(vm.OpLoadLocal, int64(slot), n.SrcPos)
	} else {
		g.emit(vm.OpLoadGlobal, int64(slot), n.SrcPos)
	}
	g.lastType = typeInt
	return nil
}

func (g *VMCodegen) VisitUnaryExpr(n *ast.UnaryExpr) error {
	if err := g.emitIntExpr(n.Operand); err != nil {
		return err
	}
	switch n.Op {
	case lex.TokenMinus:
		g.emit(vm.OpNeg, 0, n.SrcPos)
		g.lastType = typeInt
		return nil
	default:
		return newCompileError(n.SrcPos, "unsupported unary operator %d", n.Op)
	}
}

func (g *VMCodegen) VisitBinaryExpr(n *ast.BinaryExpr) error {
	if err := g.emitIntExpr(n.Left); err != nil {
		return err
	}
	if err := g.emitIntExpr(n.Right); err != nil {
		return err
	}
	switch n.Op {
	case lex.TokenPlus:
		g.emit(vm.OpAdd, 0, n.SrcPos)
	case lex.TokenMinus:
		g.emit(vm.OpSub, 0, n.SrcPos)
	case lex.TokenStar:
		g.emit(vm.OpMul, 0, n.SrcPos)
	case lex.TokenSlash:
		g.emit(vm.OpDiv, 0, n.SrcPos)
	case lex.TokenPercent:
		g.emit(vm.OpMod, 0, n.SrcPos)
	default:
		return newCompileError(n.SrcPos, "unsupported binary operator %d", n.Op)
	}
	g.lastType = typeInt
	return nil
}

func (g *VMCodegen) VisitCallExpr(n *ast.CallExpr) error {
	if n.Callee == "main" {
		return newCompileError(n.SrcPos, "explicit call to 'main' is forbidden")
	}

	// 1. Встроенные функции.
	if spec, ok := builtins[n.Callee]; ok {
		argc := len(n.Args)
		if argc < spec.minArity || (spec.maxArity >= 0 && argc > spec.maxArity) {
			if spec.maxArity < 0 {
				return newCompileError(n.SrcPos, "function %q expects at least %d argument(s), got %d", n.Callee, spec.minArity, argc)
			}
			return newCompileError(n.SrcPos, "function %q expects %d argument(s), got %d", n.Callee, spec.minArity, argc)
		}
		for _, arg := range n.Args {
			if err := g.emitIntExpr(arg); err != nil {
				return err
			}
		}
		var opArg int64
		if spec.passArgc {
			opArg = int64(argc)
		}
		g.emit(spec.op, opArg, n.SrcPos)
		g.lastType = spec.retType
		return nil
	}

	// 2. Пользовательские функции.
	sym, ok := g.funcs[n.Callee]
	if !ok {
		return newCompileError(n.SrcPos, "unknown function %q", n.Callee)
	}
	if len(n.Args) != sym.arity {
		return newCompileError(n.SrcPos, "function %q expects %d argument(s), got %d", n.Callee, sym.arity, len(n.Args))
	}
	for _, arg := range n.Args {
		if err := g.emitIntExpr(arg); err != nil {
			return err
		}
	}
	g.emit(vm.OpCall, int64(sym.index), n.SrcPos)
	g.lastType = sym.retType
	return nil
}

func (g *VMCodegen) emitIntExpr(expr ast.Expr) error {
	if err := expr.Accept(g); err != nil {
		return err
	}
	if g.lastType != typeInt {
		return newCompileError(expr.Pos(), "void function call cannot be used as a value")
	}
	return nil
}

func (g *VMCodegen) resolveVar(name string) (slot int, isLocal bool, ok bool) {
	if g.curFunc != nil {
		for _, v := range slices.Backward(g.curFunc.scopes) {
			if slot, exists := v[name]; exists {
				return slot, true, true
			}
		}
	}
	if slot, exists := g.globals[name]; exists {
		return slot, false, true
	}
	return 0, false, false
}

func blockAlwaysReturns(b *ast.BlockStmt) bool {
	for _, stmt := range b.Stmts {
		switch s := stmt.(type) {
		case *ast.ReturnStmt:
			return true
		case *ast.BlockStmt:
			if blockAlwaysReturns(s) {
				return true
			}
		}
	}
	return false
}
