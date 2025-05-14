package funcs

import (
	"testing"

	"github.com/GuanceCloud/pipeline-go/pkg/arbiter/dql"
	"github.com/GuanceCloud/platypus/pkg/engine/runtimev2"
)

func TestDQLTimeRange(t *testing.T) {
	cases := append([]ProgCase{}, cDQLTimerangeGet.Progs...)
	for _, c := range cases {
		cli := dql.NewDQLOpenAPI(
			"",
			dql.OpenAPIPath,
			"abc", nil,
		)
		if v, ok := c.privateData["time_range"]; ok {
			if v, ok := v.([]int64); ok {
				cli.Timerange = v
			}
		}
		if len(cli.Timerange) != 2 {
			end := int64(1672532100000)
			start := genTimeRange15min(end)
			cli.Timerange = []int64{start, end}
		}

		runCase(t, c, map[runtimev2.TaskP]any{
			PDQLCli: cli,
		})
	}

}
