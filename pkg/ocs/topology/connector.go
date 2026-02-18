package topology

import "time"

// AdjacencyList is the common shape of workload topology from any data provider.
// Keys are source workloads, values are destination workloads.
type AdjacencyList map[string][]string

// Connector is the interface every data-provider connector must implement.
type Connector interface {
	// Name returns the connector identifier (e.g. "istio").
	Name() string
	// FetchTopology returns workload adjacency list for the given time window.
	// If from and to are nil, the connector may use "now" or config default window.
	FetchTopology(sourceWorkloads []string, from, to *time.Time) (AdjacencyList, error)
}
