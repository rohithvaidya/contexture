package prompt

import (
	"fmt"

	"github.com/contexture/ocs/pkg/ocs/internal/config"
	connectors "github.com/contexture/ocs/pkg/ocs/topology"
)

// BuildContextDefinitions builds OCS context definitions from adjacency list and config
func BuildContextDefinitions(adjacencyList connectors.AdjacencyList, cfg *config.OCSConfig) []config.OCSContextDefinition {
	if adjacencyList == nil {
		adjacencyList = make(connectors.AdjacencyList)
	}

	workloadSet := make(map[string]bool)
	for source, destinations := range adjacencyList {
		workloadSet[source] = true
		for _, dest := range destinations {
			workloadSet[dest] = true
		}
	}
	for _, workload := range cfg.Workload {
		workloadSet[workload] = true
	}

	var out []config.OCSContextDefinition
	for workload := range workloadSet {
		contextDef := config.OCSContextDefinition{
			ResourceID: fmt.Sprintf("workload-%s", workload),
			Domain:     "compute.k8s",
			Identity: map[string]interface{}{
				"workload": workload,
			},
			Metrics: cfg.Metrics,
			Policy:  cfg.Policy,
		}

		topology := buildTopology(adjacencyList, workload)
		if len(topology) > 0 {
			contextDef.Topology = topology
		}

		out = append(out, contextDef)
	}

	return out
}

func buildTopology(adjacencyList connectors.AdjacencyList, workload string) map[string]interface{} {
	topology := make(map[string]interface{})

	if destinations, exists := adjacencyList[workload]; exists && len(destinations) > 0 {
		topology["dependencies"] = destinations
	}

	var reverseDeps []string
	for source, destinations := range adjacencyList {
		for _, dest := range destinations {
			if dest == workload {
				reverseDeps = append(reverseDeps, source)
			}
		}
	}
	if len(reverseDeps) > 0 {
		topology["dependents"] = reverseDeps
	}

	return topology
}
