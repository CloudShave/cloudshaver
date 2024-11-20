package aws

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudshave/cloudshaver/internal/pricing/client"
)

const (
	EC2Service = "AmazonEC2"
	EBSService = "AmazonEBS"
)

type EC2PricingService struct {
	client           *client.PricingClient
	supportedRegions map[string]bool
}

type ProductAttributes struct {
	InstanceType    string `json:"instanceType"`
	VCpu            string `json:"vcpu"`
	Memory          string `json:"memory"`
	Storage         string `json:"storage"`
	OperatingSystem string `json:"operatingSystem"`
	PreInstalledSw  string `json:"preInstalledSw"`
	UsageType       string `json:"usageType"`
	Operation       string `json:"operation"`
	CapacityStatus  string `json:"capacitystatus"`
	VolumeType      string `json:"volumeType"`
}

type PriceDimension struct {
	Unit         string            `json:"unit"`
	PricePerUnit map[string]string `json:"pricePerUnit"`
	Description  string            `json:"description"`
}

type TermAttributes struct {
	LeaseContractLength string `json:"LeaseContractLength"`
	PurchaseOption      string `json:"PurchaseOption"`
	OfferingClass       string `json:"OfferingClass"`
}

// NewEC2PricingService creates a new EC2 pricing service
func NewEC2PricingService(region string) (*EC2PricingService, error) {
	client := client.NewPricingClient(region)

	// Get list of supported regions
	index, err := client.GetServiceIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to get service index: %w", err)
	}

	supportedRegions := make(map[string]bool)
	if ec2Offer, ok := index.Offers[EC2Service]; ok {
		for region := range ec2Offer.Regions {
			supportedRegions[region] = true
		}
	}

	return &EC2PricingService{
		client:           client,
		supportedRegions: supportedRegions,
	}, nil
}

// IsRegionSupported checks if a region is supported for pricing
func (s *EC2PricingService) IsRegionSupported(region string) bool {
	return s.supportedRegions[region]
}

// GetInstancePrice retrieves the price for a specific EC2 instance type
func (s *EC2PricingService) GetInstancePrice(instanceType, region string) (float64, error) {
	if !s.IsRegionSupported(region) {
		return 0, fmt.Errorf("region %s is not supported for pricing", region)
	}

	data, err := s.client.GetServicePricing(EC2Service, region)
	if err != nil {
		return 0, fmt.Errorf("failed to get EC2 pricing data: %w", err)
	}

	// Parse and process pricing data
	var pricing struct {
		Products map[string]struct {
			Attributes ProductAttributes `json:"attributes"`
		} `json:"products"`
		Terms struct {
			OnDemand map[string]map[string]struct {
				PriceDimensions map[string]PriceDimension `json:"priceDimensions"`
			} `json:"OnDemand"`
		} `json:"terms"`
	}

	if err := json.Unmarshal(data, &pricing); err != nil {
		return 0, fmt.Errorf("failed to parse pricing data: %w", err)
	}

	// Find matching instance type
	for _, product := range pricing.Products {
		if product.Attributes.InstanceType == instanceType &&
			product.Attributes.OperatingSystem == "Linux" &&
			product.Attributes.PreInstalledSw == "NA" {
			// Find corresponding price
			for _, term := range pricing.Terms.OnDemand {
				for _, price := range term {
					for _, dimension := range price.PriceDimensions {
						if priceStr, ok := dimension.PricePerUnit["USD"]; ok {
							return parsePrice(priceStr)
						}
					}
				}
			}
		}
	}

	return 0, fmt.Errorf("no pricing found for instance type %s in region %s", instanceType, region)
}

// GetVolumePrice retrieves the price for a specific EBS volume type
func (s *EC2PricingService) GetVolumePrice(volumeType, region string) (float64, error) {
	data, err := s.client.GetServicePricing(EBSService, region)
	if err != nil {
		return 0, err
	}

	var pricing struct {
		Products map[string]struct {
			Attributes ProductAttributes `json:"attributes"`
		} `json:"products"`
		Terms struct {
			OnDemand map[string]map[string]struct {
				PriceDimensions map[string]PriceDimension `json:"priceDimensions"`
			} `json:"OnDemand"`
		} `json:"terms"`
	}

	if err := json.Unmarshal(data, &pricing); err != nil {
		return 0, fmt.Errorf("failed to parse EBS pricing: %v", err)
	}

	// Find the product ID for the volume type
	var productID string
	for id, product := range pricing.Products {
		attrs := product.Attributes
		if strings.EqualFold(attrs.VolumeType, volumeType) {
			productID = id
			break
		}
	}

	if productID == "" {
		return 0, fmt.Errorf("volume type %s not found in pricing data", volumeType)
	}

	// Find the price for the product
	for _, term := range pricing.Terms.OnDemand[productID] {
		for _, dimension := range term.PriceDimensions {
			if dimension.Unit == "GB-Mo" {
				for _, price := range dimension.PricePerUnit {
					return parsePrice(price)
				}
			}
		}
	}

	return 0, fmt.Errorf("no pricing found for volume type %s", volumeType)
}

// CalculateInstanceSavings calculates potential savings for an EC2 instance
func (s *EC2PricingService) CalculateInstanceSavings(currentType, targetType, region string) (float64, error) {
	currentPrice, err := s.GetInstancePrice(currentType, region)
	if err != nil {
		return 0, fmt.Errorf("failed to get current instance price: %v", err)
	}

	targetPrice, err := s.GetInstancePrice(targetType, region)
	if err != nil {
		return 0, fmt.Errorf("failed to get target instance price: %v", err)
	}

	hourlyDiff := currentPrice - targetPrice
	monthlySavings := hourlyDiff * 730 // Average hours in a month

	return monthlySavings, nil
}

// CalculateVolumeSavings calculates potential savings for an EBS volume
func (s *EC2PricingService) CalculateVolumeSavings(currentType, targetType string, sizeGB int, region string) (float64, error) {
	currentPrice, err := s.GetVolumePrice(currentType, region)
	if err != nil {
		return 0, fmt.Errorf("failed to get current volume price: %v", err)
	}

	targetPrice, err := s.GetVolumePrice(targetType, region)
	if err != nil {
		return 0, fmt.Errorf("failed to get target volume price: %v", err)
	}

	monthlySavings := float64(sizeGB) * (currentPrice - targetPrice)
	return monthlySavings, nil
}

func parsePrice(price string) (float64, error) {
	var value float64
	if _, err := fmt.Sscanf(price, "%f", &value); err != nil {
		return 0, fmt.Errorf("failed to parse price %s: %v", price, err)
	}
	return value, nil
}
