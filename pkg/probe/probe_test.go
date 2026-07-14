// pkg/probe/probe_test.go
package probe

import (
	"context"
	"net/http"
	"testing"

	"github.com/Haameed/f5_f5os_exporter/internal/config"
)

func TestProbe(t *testing.T) {
	cfg := config.F5F5osConfig{
		ScrapeTimeout: 30,
	}

	collector := &Collector{}

	_, err := collector.Probe(context.Background(), map[string]string{
		"target": "https://example.com",
	}, &http.Client{}, cfg)

	if err != nil {
		t.Logf("Expected error with invalid target: %v", err)
	}
}
