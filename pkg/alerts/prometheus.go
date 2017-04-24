package alerts

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api/prometheus"
	"github.com/prometheus/common/model"
)

// Return true if there are any active (e.g. pending or firing) alerts
func PrometheusCountActive(prometheusURL string) (int, error) {
	client, err := prometheus.New(prometheus.Config{Address: prometheusURL})
	if err != nil {
		return 0, err
	}

	queryAPI := prometheus.NewQueryAPI(client)

	value, err := queryAPI.Query(context.Background(), "ALERTS", time.Now())
	if err != nil {
		return 0, err
	}

	if value.Type() == model.ValVector {
		if vector, ok := value.(model.Vector); ok {
			// Vector contains only pending/firing alerts, no filtering required
			return len(vector), nil
		}
	}

	return 0, fmt.Errorf("Unexpected value type: %v", value)
}
