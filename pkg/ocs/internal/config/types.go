package config

// MetricConfig represents a metric configuration
type MetricConfig struct {
	Name             string                 `yaml:"name"`
	Type             string                 `yaml:"type"`
	Unit             string                 `yaml:"unit"`
	Description      string                 `yaml:"description"`
	AggregationLogic string                 `yaml:"aggregation_logic,omitempty"`
	HealthConfig     map[string]interface{} `yaml:"health_config,omitempty"`
}

// OCSConfig represents the OCS configuration structure
type OCSConfig struct {
	Policy            []string       `yaml:"policy"`
	Metrics           []MetricConfig `yaml:"metrics"`
	Workload          []string       `yaml:"workload"`
	TimeWindowMinutes *int           `yaml:"time_window_minutes"`
}

// PrometheusConfig represents Prometheus configuration
type PrometheusConfig struct {
	PrometheusInstances []struct {
		Name       string            `yaml:"name"`
		BaseURL    string            `yaml:"base_url"`
		Headers    map[string]string `yaml:"headers"`
		DisableSSL bool              `yaml:"disable_ssl"`
	} `yaml:"prometheus_instances"`
}

// OCSContextDefinition represents a context definition in the OCS prompt response
type OCSContextDefinition struct {
	ResourceID string                 `json:"resource_id,omitempty"`
	Domain     string                 `json:"domain,omitempty"`
	Identity   map[string]interface{} `json:"identity,omitempty"`
	Metrics    []MetricConfig         `json:"metrics,omitempty"`
	Topology   map[string]interface{} `json:"topology,omitempty"`
	Policy     []string               `json:"policy,omitempty"`
}

// OCSPromptResponse represents the OCS prompt response structure
type OCSPromptResponse struct {
	SpecVersion        string                 `json:"spec_version"`
	ContextDefinitions []OCSContextDefinition `json:"context_definitions"`
}
