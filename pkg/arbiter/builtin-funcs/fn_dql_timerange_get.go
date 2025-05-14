package funcs

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/pipeline-go/pkg/arbiter/dql"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtimev2"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

var FnDQLTimerangeGetDesc = runtimev2.FnDesc{
	Name:   "dql_timerange_get",
	Desc:   "Get the time range of the DQL query, which is passed in by the script caller or defaults to the last 15 minutes.",
	Params: []*runtimev2.Param{},
	Returns: []*runtimev2.Param{
		{
			Desc: "The time range. For example, `[1744214400000, 1744218000000]`, the timestamp precision is milliseconds",
			Typs: []ast.DType{ast.List},
		},
	},
}

func FnDQLTimerangeGetCheck(ctx *runtimev2.Task, expr *ast.CallExpr) *errchain.PlError {
	return runtimev2.CheckPassParam(ctx, expr, FnDQLTimerangeGetDesc.Params)
}

func FnDQLTimerangeGet(ctx *runtimev2.Task, expr *ast.CallExpr) *errchain.PlError {
	v, ok := ctx.PValue(PDQLCli)
	if !ok {
		return runtimev2.NewRunError(ctx, fmt.Sprintf(
			"missing context data named %s", PDQLCli), expr.NamePos)
	}
	dqlCli, ok := v.(dql.DQL)
	if !ok {
		return runtimev2.NewRunError(ctx, fmt.Sprintf("context data %s type is expected", PDQLCli), expr.NamePos)
	}
	r := dqlCli.TimeRange()
	if len(r) != 2 {
		end := time.Now().UnixMilli()
		start := genTimeRange15min(end)
		ctx.Regs.ReturnAppend(runtimev2.V{
			V: []any{start, end},
			T: ast.List,
		})
	} else {
		ctx.Regs.ReturnAppend(runtimev2.V{
			V: []any{r[0], r[1]},
			T: ast.List,
		})
	}
	return nil
}

func genTimeRange15min(end int64) int64 {
	return end - int64(time.Minute)/int64(time.Millisecond)*15
}
