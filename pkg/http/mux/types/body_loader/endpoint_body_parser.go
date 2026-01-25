package body_loader

import (
	"github.com/Motmedel/utils_go/pkg/http/mux/types/body_loader/body_setting"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/body_parser"
)

type Loader struct {
	Parser      body_parser.BodyParser[any]
	ContentType string
	Setting     body_setting.Setting
	MaxBytes    int64
}
