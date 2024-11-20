package aws

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// RDSPricing holds pricing information for RDS instances
type RDSPricing struct {
	RegionMapping map[string]map[string]RDSInstancePricing
	dataDir       string
}

// RDSInstancePricing represents pricing information for an RDS instance type
type RDSInstancePricing struct {
	OnDemandPrice float64 `json:"onDemandPrice"`
}

// NewRDSPricing creates a new RDSPricing instance
func NewRDSPricing(dataDir string) *RDSPricing {
	return &RDSPricing{
		RegionMapping: make(map[string]map[string]RDSInstancePricing),
		dataDir:       dataDir,
	}
}

// LoadPricing loads RDS pricing data from JSON files
func (p *RDSPricing) LoadPricing() error {
	// Load instance pricing
	data, err := os.ReadFile(filepath.Join(p.dataDir, "internal", "pricing", "aws", "data", "rds_pricing.json"))
	if err != nil {
		return fmt.Errorf("failed to read RDS pricing data: %v", err)
	}

	if err := json.Unmarshal(data, &p.RegionMapping); err != nil {
		return fmt.Errorf("failed to parse RDS pricing data: %v", err)
	}

	return nil
}

// IsRegionSupported checks if pricing is supported for the given region
func (p *RDSPricing) IsRegionSupported(region string) bool {
	_, ok := p.RegionMapping[region]
	return ok
}

// CalculateInstanceSavings calculates potential savings when upgrading from one instance type to another
func (p *RDSPricing) CalculateInstanceSavings(currentType, targetType, region string) (float64, error) {
	if !p.IsRegionSupported(region) {
		return 0, fmt.Errorf("region %s is not supported", region)
	}

	currentPricing, ok := p.RegionMapping[region][currentType]
	if !ok {
		return 0, fmt.Errorf("no pricing data available for instance type %s", currentType)
	}

	targetPricing, ok := p.RegionMapping[region][targetType]
	if !ok {
		return 0, fmt.Errorf("no pricing data available for instance type %s", targetType)
	}

	// Calculate monthly savings (assuming 720 hours per month)
	monthlySavings := (currentPricing.OnDemandPrice - targetPricing.OnDemandPrice) * 720
	return monthlySavings, nil
}
