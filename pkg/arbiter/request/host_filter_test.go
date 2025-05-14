package request

import (
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHostFilte(t *testing.T) {
	cases := []struct {
		name         string
		blockdCIDRs  []string
		allowedCIDRs []string
		blockdHosts  []string
		allowedHosts []string

		addr    string
		blocked bool
	}{
		{
			name: "test",
			addr: "baidu.com:80",
		},
		{
			name:        "test",
			addr:        "baidu.com:80",
			blockdHosts: []string{"abc.com", "baidu.com", "qq.com"},
			blocked:     true,
		},
		{
			name:         "test",
			addr:         "baidu.com:80",
			allowedHosts: []string{"abc.com", "qq.com"},
			blocked:      true,
		},
		{
			name:         "test",
			addr:         "baidu.com:80",
			allowedCIDRs: []string{"1.0.0.0/8"},
			blocked:      true,
		},
		{
			name:         "test",
			addr:         "baidu.com:80",
			allowedCIDRs: []string{"39.0.0.0/8", "110.0.0.0/8"},
		},
		{
			name:         "test",
			addr:         "baidu.com:80",
			allowedCIDRs: []string{"39.0.0.0/8", "110.0.0.0/8"},
		},
		{
			name:        "test",
			addr:        "local.ubwbu.com:1234",
			blockdCIDRs: PrivateCIDRs(),
			blocked:     true,
		},
		{
			name:        "test",
			addr:        "127.0.0.1:80",
			blockdCIDRs: PrivateCIDRs(),
			blocked:     true,
		},
	}

	for i, c := range cases {
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) {
			host, _, err := net.SplitHostPort(c.addr)
			if err != nil {
				t.Fatal(err)
			}

			f := NewHostFilter(c.blockdCIDRs, c.allowedCIDRs, c.blockdHosts, c.allowedHosts, 100, 1*time.Second)

			blocked, _ := f.IsBlocked(host)
			assert.Equal(t, c.blocked, blocked)
		})
	}
}
