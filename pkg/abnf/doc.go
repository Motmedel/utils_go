// Package abnf provides ABNF (RFC 5234, with the RFC 7405 char-val
// extension) grammar parsing and input matching.
//
// ParseABNF parses an ABNF grammar definition into a Grammar, and Parse
// matches an input against a rule of such a grammar, producing match Paths.
//
// The "#" list operator of RFC 9110 Section 5.6.1 is read as well. RFC 5234
// does not define it; the HTTP specifications do, and they give it in terms
// of standard ABNF, which is what a grammar using it is read as. The
// expansion separates elements with OWS, so a grammar using the operator
// has to define OWS, just as it has to define every other rule it refers
// to. See list.go for which of the expansions of RFC 7230 Section 7 apply
// and why.
//
// The package is derived from github.com/pandatix/go-abnf v0.4.2,
// Copyright (c) 2024 Lucas TESSON - PandatiX, licensed under the MIT
// License (see the LICENSE file in this directory), restructured for this
// module and reduced to grammar parsing and input matching.
package abnf
