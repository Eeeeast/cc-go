package lexer

type Symbol uint32

type Interner struct {
	strings []string          // ID -> string
	lookup  map[string]Symbol // string -> ID
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
	return in.strings[sym]
}
