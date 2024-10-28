package utils

import (
	"fmt"
	"strings"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func LiteralTypeToStr(lt *core.LiteralType) string {
	if lt == nil {
		return "None"
	}
	if lt.GetSimple() == core.SimpleType_STRUCT {
		var structure string
		for k, v := range lt.GetStructure().GetDataclassType() {
			structure += fmt.Sprintf("dataclass_type:{key:%v value:{%v}, ", k, LiteralTypeToStr(v))
		}
		structure = strings.TrimSuffix(structure, ", ")
		return fmt.Sprintf("Simple: STRUCT structure{%v}", structure)
	}
	return lt.String()
}
