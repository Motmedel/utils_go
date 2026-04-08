package etag

import (
	_ "embed"
	"errors"
	"fmt"

	"github.com/Motmedel/parsing_utils/pkg/parsing_utils"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	goabnf "github.com/pandatix/go-abnf"
)

//go:embed grammar.abnf
var grammar []byte

var Grammar *goabnf.Grammar

var ErrNilOpaqueTagPath = errors.New("nil opaque-tag path")

func Parse(data []byte) (*motmedelHttpTypes.ETag, error) {
	paths, err := parsing_utils.GetParsedDataPaths(Grammar, data)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("get parsed data paths: %w", err), data)
	}
	if len(paths) == 0 {
		return nil, motmedelErrors.NewWithTrace(motmedelErrors.ErrSyntaxError, data)
	}

	var etag motmedelHttpTypes.ETag

	if weakPath := parsing_utils.SearchPathSingleName(paths[0], "weak", 2, false); weakPath != nil {
		etag.Weak = true
	}

	opaqueTagPath := parsing_utils.SearchPathSingleName(paths[0], "opaque-tag", 2, false)
	if opaqueTagPath == nil {
		return nil, motmedelErrors.NewWithTrace(
			fmt.Errorf("%w: %w", motmedelErrors.ErrSemanticError, ErrNilOpaqueTagPath),
			data,
		)
	}

	opaqueTag := parsing_utils.ExtractPathValue(data, opaqueTagPath)
	if len(opaqueTag) >= 2 {
		etag.Tag = string(opaqueTag[1 : len(opaqueTag)-1])
	}

	return &etag, nil
}

func init() {
	var err error
	Grammar, err = goabnf.ParseABNF(grammar)
	if err != nil {
		panic(fmt.Sprintf("goabnf parse abnf (etag grammar): %v", err))
	}
}
