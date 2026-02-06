package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ToolInfo describes a tool's tier and service.
type ToolInfo struct {
	Tier    string
	Service string
}

// TierConfig holds the tier configuration loaded from tool_tiers.yaml.
type TierConfig struct {
	Services map[string]ServiceTiers `yaml:"services"`
}

// ServiceTiers lists tools by tier within a service.
type ServiceTiers struct {
	Core     []string `yaml:"core"`
	Extended []string `yaml:"extended"`
	Complete []string `yaml:"complete"`
}

// LoadTiers reads and parses the tool tiers YAML file, returning a map of
// tool name -> ToolInfo for fast lookup during tool filtering.
func LoadTiers(path string) (map[string]ToolInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading tier config %s: %w", path, err)
	}

	var tc TierConfig
	if err := yaml.Unmarshal(data, &tc); err != nil {
		return nil, fmt.Errorf("parsing tier config %s: %w", path, err)
	}

	tools := make(map[string]ToolInfo)
	for service, tiers := range tc.Services {
		for _, name := range tiers.Core {
			tools[name] = ToolInfo{Tier: "core", Service: service}
		}
		for _, name := range tiers.Extended {
			tools[name] = ToolInfo{Tier: "extended", Service: service}
		}
		for _, name := range tiers.Complete {
			tools[name] = ToolInfo{Tier: "complete", Service: service}
		}
	}

	return tools, nil
}

// TierLevel returns the numeric level for a tier name (higher = more inclusive).
func TierLevel(tier string) int {
	switch tier {
	case "core":
		return 1
	case "extended":
		return 2
	case "complete":
		return 3
	default:
		return 0
	}
}
