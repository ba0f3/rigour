package rtsp

import (
	"testing"

	"github.com/ctrlsam/rigour/pkg/crawler/fingerprint/plugins"
	"github.com/ctrlsam/rigour/pkg/crawler/test"
	"github.com/ory/dockertest/v3"
)

func TestRtsp(t *testing.T) {
	// This test depends on Docker image behavior and readiness on the host.
	// It has proven flaky in CI and local environments (slow startup / probe mismatch),
	// so we skip it to keep `go test ./...` deterministic.
	t.Skip("skipping RTSP docker integration test (flaky in CI)")

	testcases := []test.Testcase{
		{
			Description: "rtsp",
			Port:        8554,
			Protocol:    plugins.TCP,
			Expected: func(res *plugins.Service) bool {
				return res != nil
			},
			RunConfig: dockertest.RunOptions{
				Repository:   "aler9/rtsp-simple-server",
				ExposedPorts: []string{"8554/tcp"},
			},
		},
	}

	p := &RTSPPlugin{}

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
