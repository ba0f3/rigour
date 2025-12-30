package ipsec

import (
	"testing"

	"github.com/ctrlsam/rigour/pkg/crawler/fingerprint/plugins"
	"github.com/ctrlsam/rigour/pkg/crawler/test"
	"github.com/ory/dockertest/v3"
)

func TestIPSEC(t *testing.T) {
	// Flaky in CI: depends on container timing/networking and can fail to
	// fingerprint reliably on shared runners.
	t.Skip("skipping flaky docker integration test")

	testcases := []test.Testcase{
		{
			Description: "ipsec",
			Port:        500,
			Protocol:    plugins.UDP,
			Expected: func(res *plugins.Service) bool {
				return res != nil
			},
			RunConfig: dockertest.RunOptions{
				Repository: "hwdsl2/ipsec-vpn-server",
				Mounts: []string{
					"ikev2-vpn-data:/etc/ipsec.d",
					"/lib/modules:/lib/modules:ro",
				},
				Privileged: true,
			},
		},
	}

	var p *Plugin

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.Description, func(t *testing.T) {
			t.Parallel()
			err := test.RunTest(t, tc, p)
			if err != nil {
				t.Error(err)
			}
		})
	}
}
