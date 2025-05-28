package funcs

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	dffunc "github.com/GuanceCloud/pipeline-go/pkg/arbiter/df-func"
	"github.com/GuanceCloud/platypus/pkg/engine/runtimev2"
)

func mockDFFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var v map[string]any
	_ = json.NewDecoder(r.Body).Decode(&v)
	kwargs := v["kwargs"].(map[string]any)
	if _, ok := kwargs["arg_1"]; ok {
		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte(`{"ok":true,"error":200,"message":"","data":{"result":{"arg_1":"1","arg_2":[12,3],"kwargs":{"key3":"true"}}},"traceId":"TRACE-D8591EB8-4B94-4604-921F-69106542B461","clientTime":null,"reqTime":"2025-05-28T02:29:59.110Z","respTime":"2025-05-28T02:29:59.294Z","reqCost":184}`))
	} else {
		w.WriteHeader(599)
		_, _ = w.Write([]byte(`{"ok":false,"error":599.1,"reason":"EFuncFailed","message":"Func task failed","detail":{"exception":"TypeError(\"echo() missing 1 required positional argument: 'arg_1'\")","traceback":"Traceback (most recent call last *in User Script*):\nTypeError: echo() missing 1 required positional argument: 'arg_1'"},"status":599,"reqDump":{"method":"POST","url":"/api/v1/func/guance_siem__api.echo","bodyDump":"{\n  \"kwargs\": {\n    \"arg_2\": [\n      12,\n      3\n    ]\n  }\n}"},"traceId":"TRACE-55F0EDEA-4C86-4135-9E26-4971C1222880","clientTime":null,"reqTime":"2025-05-28T02:29:59.322Z","respTime":"2025-05-28T02:29:59.497Z","reqCost":175}`))
	}
}

func TestCallFunc(t *testing.T) {
	cases := []ProgCase{}
	cases = append(cases, cCallFunc.Progs...)

	server := httptest.NewServer(http.HandlerFunc(mockDFFunc))
	defer server.Close()

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			runCase(t, tc, map[runtimev2.TaskP]any{
				PCallFunc: dffunc.NewDFFunc(server.URL),
			})
		})
	}
}
