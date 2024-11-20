package aws

import (
	"fmt"
	awsinterfaces "github.com/cloudshave/cloudshaver/internal/aws/interfaces"
)

// PricingService implements the PricingServiceAPI interface
type PricingService struct {
	pricing *EC2Pricing
}

// NewPricingService creates a new AWS pricing service
func NewPricingService() (awsinterfaces.PricingServiceAPI, error) {
	pricing, err := LoadPricing()
	if err != nil {
		return nil, fmt.Errorf("failed to load pricing data: %v", err)
	}
	return &PricingService{
		pricing: pricing,
	}, nil
}

// IsRegionSupported checks if pricing is supported for the given region
func (p *PricingService) IsRegionSupported(region string) bool {
	_, ok := p.pricing.RegionMapping[region]
	return ok
}

// GetVolumePrice returns the price per GB-month for the given volume type in the specified region
func (p *PricingService) GetVolumePrice(volumeType, region string) (float64, error) {
	if !p.IsRegionSupported(region) {
		return 0, fmt.Errorf("region %s is not supported", region)
	}

	regionPricing, ok := p.pricing.EBSVolumes[region]
	if !ok {
		return 0, fmt.Errorf("no pricing data available for region %s", region)
	}

	volume, ok := regionPricing[volumeType]
	if !ok {
		return 0, fmt.Errorf("no pricing data available for volume type %s", volumeType)
	}

	return volume.PricePerGBMonth, nil
}

// CalculateInstanceSavings calculates potential savings when upgrading from one instance type to another
func (p *PricingService) CalculateInstanceSavings(currentType, targetType, region string) (float64, error) {
	savings, _, err := p.pricing.CalculateInstanceSavings(region, currentType, 720) // Assume 30 days of runtime
	if err != nil {
		return 0, err
	}
	return savings, nil
}
