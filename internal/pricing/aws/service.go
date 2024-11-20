package aws

import (
	"fmt"
	"os"
	"path/filepath"
	awsinterfaces "github.com/cloudshave/cloudshaver/internal/interfaces/aws"
)

// PricingService implements the PricingServiceAPI interface
type PricingService struct {
	ec2Pricing *EC2Pricing
	rdsPricing *RDSPricing
}

// NewPricingService creates a new instance of PricingService
func NewPricingService() (awsinterfaces.PricingServiceAPI, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %v", err)
	}
	execDir := filepath.Dir(execPath)

	service := &PricingService{
		ec2Pricing: NewEC2Pricing(execDir),
		rdsPricing: NewRDSPricing(execDir),
	}

	if err := service.LoadPricing(); err != nil {
		return nil, fmt.Errorf("failed to load pricing data: %v", err)
	}

	return service, nil
}

// LoadPricing loads pricing data for both EC2 and RDS
func (s *PricingService) LoadPricing() error {
	if err := s.ec2Pricing.LoadPricing(); err != nil {
		return fmt.Errorf("failed to load EC2 pricing: %v", err)
	}
	if err := s.rdsPricing.LoadPricing(); err != nil {
		return fmt.Errorf("failed to load RDS pricing: %v", err)
	}
	return nil
}

// CalculateInstanceSavings calculates potential savings for an instance
func (s *PricingService) CalculateInstanceSavings(currentType, targetType, region string) (float64, error) {
	// Try EC2 pricing first
	savings, err := s.ec2Pricing.CalculateInstanceSavings(currentType, targetType, region)
	if err == nil {
		return savings, nil
	}

	// If EC2 pricing fails, try RDS pricing
	savings, err = s.rdsPricing.CalculateInstanceSavings(currentType, targetType, region)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate savings: %v", err)
	}

	return savings, nil
}

// IsRegionSupported checks if the given region is supported
func (s *PricingService) IsRegionSupported(region string) bool {
	return s.ec2Pricing.IsRegionSupported(region) || s.rdsPricing.IsRegionSupported(region)
}

// GetVolumePrice returns the price per GB-month for the given volume type in the specified region
func (s *PricingService) GetVolumePrice(volumeType, region string) (float64, error) {
	if !s.IsRegionSupported(region) {
		return 0, fmt.Errorf("region %s is not supported", region)
	}

	regionPricing, ok := s.ec2Pricing.EBSVolumes[region]
	if !ok {
		return 0, fmt.Errorf("no pricing data available for region %s", region)
	}

	volume, ok := regionPricing[volumeType]
	if !ok {
		return 0, fmt.Errorf("no pricing data available for volume type %s", volumeType)
	}

	return volume.PricePerGBMonth, nil
}
