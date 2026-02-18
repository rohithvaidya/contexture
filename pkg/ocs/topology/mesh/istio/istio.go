package istio

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/contexture/ocs/pkg/ocs/telemetry"
	"github.com/contexture/ocs/pkg/ocs/topology"
)

type IstioConnector struct {
	prometheusURL string
	httpClient    *http.Client
}

// NewConnector creates a new Istio connector.
func Create(prometheusURL string) *IstioConnector {
	return &IstioConnector{
		prometheusURL: prometheusURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name implements topology.Connector.
func (c *IstioConnector) Name() string {
	return "istio"
}

// FetchTopology implements topology.Connector.
// It queries Prometheus for istio_requests_total and returns workload adjacency list.
func (c *IstioConnector) FetchTopology(sourceWorkloads []string, from, to *time.Time) (topology.AdjacencyList, error) {
	if len(sourceWorkloads) == 0 {
		return nil, fmt.Errorf("no source workloads provided")
	}

	workloadFilter := strings.Join(sourceWorkloads, "|")
	query := fmt.Sprintf(`istio_requests_total{source_workload=~"%s"}`, workloadFilter)

	var result *telemetry.PrometheusQueryResult
	var err error
	if from != nil && to != nil {
		result, err = c.queryRange(query, from, to)
	} else {
		result, err = c.queryInstant(query)
	}
	if err != nil {
		return nil, err
	}

	return extractAdjacencyList(result), nil
}

func (c *IstioConnector) queryRange(query string, from, to *time.Time) (*telemetry.PrometheusQueryResult, error) {
	start := from.Unix()
	end := to.Unix()
	step := "15s"

	queryURL := fmt.Sprintf("%s/api/v1/query_range?query=%s&start=%d&end=%d&step=%s",
		c.prometheusURL, url.QueryEscape(query), start, end, step)
	log.Printf("Querying Prometheus (range): %s from %s to %s", query, from.Format(time.RFC3339), to.Format(time.RFC3339))

	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Prometheus returned status %d: %s", resp.StatusCode, string(body))
	}

	var rangeResult *telemetry.PrometheusQueryRangeResult
	if err := json.NewDecoder(resp.Body).Decode(&rangeResult); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if rangeResult.Status != "success" {
		return nil, fmt.Errorf("Prometheus query failed with status: %s", rangeResult.Status)
	}

	instantResult := convertRangeToInstantResult(rangeResult)
	log.Printf("Retrieved %d unique metrics from Prometheus range query", len(instantResult.Data.Result))
	return instantResult, nil
}

func (c *IstioConnector) queryInstant(query string) (*telemetry.PrometheusQueryResult, error) {
	queryURL := fmt.Sprintf("%s/api/v1/query?query=%s", c.prometheusURL, url.QueryEscape(query))
	log.Printf("Querying Prometheus (instant): %s", query)

	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Prometheus returned status %d: %s", resp.StatusCode, string(body))
	}

	var result telemetry.PrometheusQueryResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("Prometheus query failed with status: %s", result.Status)
	}

	log.Printf("Retrieved %d results from Prometheus", len(result.Data.Result))
	return &result, nil
}

func convertRangeToInstantResult(r *telemetry.PrometheusQueryRangeResult) *telemetry.PrometheusQueryResult {
	instantResult := &telemetry.PrometheusQueryResult{Status: r.Status}

	uniqueMetrics := make(map[string]struct {
		Metric map[string]string
		Seen   bool
	})

	for _, res := range r.Data.Result {
		metricKey := fmt.Sprintf("%v", res.Metric)
		if _, exists := uniqueMetrics[metricKey]; !exists {
			uniqueMetrics[metricKey] = struct {
				Metric map[string]string
				Seen   bool
			}{Metric: res.Metric, Seen: true}
		}
	}

	for _, v := range uniqueMetrics {
		instantResult.Data.Result = append(instantResult.Data.Result, struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
		}{
			Metric: v.Metric,
			Value:  []interface{}{time.Now().Unix(), "1"},
		})
	}

	return instantResult
}

func extractAdjacencyList(result *telemetry.PrometheusQueryResult) topology.AdjacencyList {
	adjacencyList := make(topology.AdjacencyList)

	for _, r := range result.Data.Result {
		source := r.Metric["source_workload"]
		destination := r.Metric["destination_workload"]

		if source != "" && destination != "" {
			if adjacencyList[source] == nil {
				adjacencyList[source] = make([]string, 0)
			}
			exists := false
			for _, dest := range adjacencyList[source] {
				if dest == destination {
					exists = true
					break
				}
			}
			if !exists {
				adjacencyList[source] = append(adjacencyList[source], destination)
			}
		}
	}

	log.Printf("Extracted adjacency list with %d sources", len(adjacencyList))
	return adjacencyList
}
