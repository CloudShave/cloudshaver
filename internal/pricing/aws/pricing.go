package aws

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// EC2Pricing holds pricing information for EC2 instances and EBS volumes
type EC2Pricing struct {
	RegionMapping map[string]map[string]InstancePricing
	EBSVolumes    map[string]map[string]VolumePricing
	dataDir       string
}

// InstancePricing represents pricing information for an EC2 instance type
type InstancePricing struct {
	OnDemandPrice float64 `json:"onDemandPrice"`
}

// VolumePricing represents pricing information for an EBS volume type
type VolumePricing struct {
	PricePerGBMonth float64 `json:"pricePerGBMonth"`
}

// NewEC2Pricing creates a new EC2Pricing instance
func NewEC2Pricing(dataDir string) *EC2Pricing {
	return &EC2Pricing{
		RegionMapping: make(map[string]map[string]InstancePricing),
		EBSVolumes:    make(map[string]map[string]VolumePricing),
		dataDir:       dataDir,
	}
}

// LoadPricing loads EC2 pricing data from JSON files
func (p *EC2Pricing) LoadPricing() error {
	// Load instance pricing
	instanceData, err := os.ReadFile(filepath.Join(p.dataDir, "data", "ec2_pricing.json"))
	if err != nil {
		return fmt.Errorf("failed to read EC2 pricing data: %v", err)
	}

	if err := json.Unmarshal(instanceData, &p.RegionMapping); err != nil {
		return fmt.Errorf("failed to parse EC2 pricing data: %v", err)
	}

	// Load EBS volume pricing
	volumeData, err := os.ReadFile(filepath.Join(p.dataDir, "data", "ebs_pricing.json"))
	if err != nil {
		return fmt.Errorf("failed to read EBS pricing data: %v", err)
	}

	if err := json.Unmarshal(volumeData, &p.EBSVolumes); err != nil {
		return fmt.Errorf("failed to parse EBS pricing data: %v", err)
	}

	return nil
}

// IsRegionSupported checks if pricing is supported for the given region
func (p *EC2Pricing) IsRegionSupported(region string) bool {
	_, ok := p.RegionMapping[region]
	return ok
}

// CalculateInstanceSavings calculates potential savings when upgrading from one instance type to another
func (p *EC2Pricing) CalculateInstanceSavings(currentType, targetType, region string) (float64, error) {
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
