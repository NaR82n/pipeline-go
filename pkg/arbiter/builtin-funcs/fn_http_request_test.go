package funcs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/GuanceCloud/pipeline-go/pkg/arbiter/request"
	"github.com/GuanceCloud/platypus/pkg/engine/runtimev2"
)

func mockHTTPResp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"code":200, "message":"success"}`))
}

func TestHTTPRequest(t *testing.T) {
	cases := []ProgCase{}
	server := httptest.NewServer(http.HandlerFunc(mockHTTPResp))
	defer server.Close()

	cases = append(cases, cHTTPRequest.Progs...)
	for _, tc := range cases {
		replaceStr := tc.privateData["replace"].(string)
		var blockedCIDRs []string
		if v, ok := tc.privateData["blockedCIDRs"]; ok {
			blockedCIDRs = v.([]string)
		}
		tc.Script = strings.Replace(tc.Script, replaceStr, server.URL, -1)
		tc.Stdout = strings.Replace(tc.Stdout, replaceStr, server.URL, -1)
		request.SetTransport(request.NewFilteringDialer(request.NewHostFilter(
			blockedCIDRs,
			nil,
			nil,
			nil,
			100, time.Second*5,
		)))

		runCase(t, tc, map[runtimev2.TaskP]any{
			PHTTPClient: request.NewHTTPClient(0),
		})
	}
}
