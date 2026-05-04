package spf

import (
	_ "embed"
	"fmt"

	goabnf "github.com/pandatix/go-abnf"
)

//go:embed grammar.abnf
var grammar []byte

var SpfGrammar *goabnf.Grammar

func init() {
	var err error
	SpfGrammar, err = goabnf.ParseABNF(grammar)
	if err != nil {
		panic(fmt.Sprintf("goabnf parse abnf: %v", err))
	}
}
