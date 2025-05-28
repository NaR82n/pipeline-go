package funcs

import (
	"fmt"

	dffunc "github.com/GuanceCloud/pipeline-go/pkg/arbiter/df-func"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtimev2"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

var FnCallFuncDesc = runtimev2.FnDesc{
	Name: "call_func",
	Desc: "Calling remote functions on the Function platform",
	Params: []*runtimev2.Param{
		{
			Name: "name",
			Desc: "Remote function name.",
			Typs: []ast.DType{ast.String},
		},
		{
			Name: "kwargs",
			Desc: "Parameter map, corresponding to **kwargs in Python.",
			Typs: []ast.DType{ast.Map},
			Val: func() any {
				return map[string]any{}
			},
		},
	},
	Returns: []*runtimev2.Param{
		{
			Typs: []ast.DType{ast.Map},
		},
	},
}

func FnCallFuncCheck(ctx *runtimev2.Task, expr *ast.CallExpr) *errchain.PlError {
	return runtimev2.CheckPassParam(ctx, expr, FnCallFuncDesc.Params)
}

func FnCallFunc(ctx *runtimev2.Task, expr *ast.CallExpr) *errchain.PlError {
	v, ok := ctx.PValue(PCallFunc)
	if !ok {
		return runtimev2.NewRunError(ctx, fmt.Sprintf("missing context data named %s", PCallFunc), expr.NamePos)
	}
	callFnCli, ok := v.(dffunc.DFCall)
	if !ok || callFnCli == nil {
		return runtimev2.NewRunError(ctx, fmt.Sprintf("context data %s type is nexpected", PCallFunc), expr.NamePos)
	}

	name, err := runtimev2.GetParamString(ctx, expr, FnCallFuncDesc.Params, 0)
	if err != nil {
		return err
	}
	kwargs, err := runtimev2.GetParamMap(ctx, expr, FnCallFuncDesc.Params, 1)
	if err != nil {
		return err
	}

	resutlt := callFnCli.Call(name, kwargs)
	ctx.Regs.ReturnAppend(runtimev2.V{
		V: resutlt,
		T: ast.Map,
	})

	return nil
}
