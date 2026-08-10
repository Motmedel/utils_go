# Notices

This repository as a whole, including all nested Go modules (`pkg/dns/parsing`,
`pkg/http/mux`, `pkg/http/parsing/headers`, ...), is licensed under the MIT
License; see [LICENSE](LICENSE). The portions listed below are derived from
third-party work and additionally carry their original licenses, which are
retained in the indicated directories.

## `pkg/abnf`

Derived from go-abnf (github.com/pandatix/go-abnf) v0.4.2,
Copyright (c) 2024 Lucas TESSON - PandatiX, licensed under the MIT License;
see [pkg/abnf/LICENSE](pkg/abnf/LICENSE). The code has been restructured
and reduced to grammar parsing and input matching. Does not apply to the
`pkg/abnf/utils` subdirectory, which is original work.

## `pkg/json/schema`

Portions are derived from the Go Authors' JSON Schema implementation
(github.com/google/jsonschema-go). Files derived from that work carry a
`Copyright 2025 The Go Authors` header and are licensed under a BSD-3-Clause
license; see [pkg/json/schema/LICENSE](pkg/json/schema/LICENSE).

## `pkg/json/schema/internal/tests`

The test data under `tests/draft2020-12` and `remotes` comes from the JSON
Schema Test Suite (github.com/json-schema-org/JSON-Schema-Test-Suite),
Copyright (c) 2012 Julian Berman, licensed under the MIT License; see the
`LICENSE` file in each of those directories.
