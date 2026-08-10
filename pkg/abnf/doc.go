// Package abnf provides ABNF (RFC 5234, with the RFC 7405 char-val
// extension) grammar parsing and input matching.
//
// ParseABNF parses an ABNF grammar definition into a Grammar, and Parse
// matches an input against a rule of such a grammar, producing match Paths.
//
// The package is derived from github.com/pandatix/go-abnf v0.4.2,
// Copyright (c) 2024 Lucas TESSON - PandatiX, licensed under the MIT
// License (see the LICENSE file in this directory), restructured for this
// module and reduced to grammar parsing and input matching.
package abnf
