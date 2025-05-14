package funcs

import (
	"fmt"

	"github.com/GuanceCloud/pipeline-go/pkg/arbiter/request"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtimev2"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

var FnHTTPRequestDesc = runtimev2.FnDesc{
	Name: "http_request",
	Desc: "Used to send http request.",
	Params: []*runtimev2.Param{
		{
			Name: "method",
			Desc: "HTTP request method",
			Typs: []ast.DType{ast.String},
		},
		{
			Name: "url",
			Desc: "HTTP request url",
			Typs: []ast.DType{ast.String},
		},
		{
			Name: "body",
			Desc: "HTTP request body",
			Typs: ast.AllTyp(),
			Val: func() any {
				return nil
			},
		},
		{
			Name: "headers",
			Desc: "HTTP request headers",
			Typs: []ast.DType{ast.Map},
			Val: func() any {
				return map[string]any{}
			},
		},
	},
	Returns: []*runtimev2.Param{
		{
			Desc: "HTTP response",
			Typs: []ast.DType{ast.Map},
		},
	},
}

func FnHTTPRequestChecking(ctx *runtimev2.Task, expr *ast.CallExpr) *errchain.PlError {
	return runtimev2.CheckPassParam(ctx, expr, FnHTTPRequestDesc.Params)
}

func FnHTTPRequest(ctx *runtimev2.Task, expr *ast.CallExpr) *errchain.PlError {
	v, ok := ctx.PValue(PHTTPClient)
	if !ok {
		return runtimev2.NewRunError(ctx, fmt.Sprintf(
			"missing context data named %s", PDQLCli), expr.NamePos)
	}
	httpCLi, ok := v.(request.HTTPCli)
	if !ok || httpCLi == nil {
		return runtimev2.NewRunError(ctx, fmt.Sprintf("context data %s type is expected", PDQLCli), expr.NamePos)
	}

	method, err := runtimev2.GetParamString(ctx, expr, FnHTTPRequestDesc.Params, 0)
	if err != nil {
		return err
	}
	url, err := runtimev2.GetParamString(ctx, expr, FnHTTPRequestDesc.Params, 1)
	if err != nil {
		return err
	}
	body, err := runtimev2.GetParam(ctx, expr, FnHTTPRequestDesc.Params, 2)
	if err != nil {
		return err
	}
	headers, err := runtimev2.GetParamMap(ctx, expr, FnHTTPRequestDesc.Params, 3)
	if err != nil {
		return err
	}

	var result map[string]any
	if r, errReq := httpCLi.Request(method, url, headers, body); errReq != nil {
		result = map[string]any{
			"error": errReq.Error(),
		}
	} else {
		result = r
	}

	ctx.Regs.ReturnAppend(runtimev2.V{
		V: result,
		T: ast.Map,
	})
	return nil
}
