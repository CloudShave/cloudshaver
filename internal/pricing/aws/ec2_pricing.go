package aws

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/cloudshave/cloudshaver/internal/pricing/client"
)

const (
	EC2Service = "AmazonEC2"
	EBSService = "AmazonEBS"
)

// Constants for different pricing scenarios
const (
	// Operating Systems
	OSLinux   = "Linux"
	OSWindows = "Windows"
	OSRHEL    = "RHEL"
	OSSuse    = "SUSE"
	
	// Tenancy Types
	TenancyShared    = "Shared"
	TenancyDedicated = "Dedicated"
	TenancyHost      = "Host"
	
	// License Models
	LicenseIncluded  = "License included"
	LicenseNoLicense = "No License required"
	LicenseBYOL      = "Bring your own license"
	
	// Capacity Status
	CapacityUsed     = "Used"
	CapacityReserved = "Reserved"
	CapacitySpot     = "Spot"
	
	// Purchase Options
	PurchaseOnDemand = "On Demand"
	PurchaseReserved = "Reserved"
	PurchaseSpot     = "Spot"
)

type EC2PricingService struct {
	client           *client.PricingClient
	supportedRegions map[string]bool
}

type ProductAttributes struct {
	// Basic Instance Attributes
	InstanceType     string `json:"instanceType"`
	VCpu            string `json:"vcpu"`
	Memory          string `json:"memory"`
	Storage         string `json:"storage"`
	
	// Operating System and Software
	OperatingSystem string `json:"operatingSystem"`
	PreInstalledSw  string `json:"preInstalledSw"`
	LicenseModel    string `json:"licenseModel"`
	
	// Usage and Capacity
	UsageType       string `json:"usageType"`
	Operation       string `json:"operation"`
	CapacityStatus  string `json:"capacitystatus"`
	
	// Location and Tenancy
	Tenancy         string `json:"tenancy"`
	Location        string `json:"location"`
	LocationType    string `json:"locationType"`
	
	// Hardware Specifications
	ProcessorArchitecture    string `json:"processorArchitecture"`
	ProcessorFeatures       string `json:"processorFeatures"`
	PhysicalProcessor      string `json:"physicalProcessor"`
	ClockSpeed            string `json:"clockSpeed"`
	
	// Network
	NetworkPerformance           string `json:"networkPerformance"`
	NetworkBandwidthGbps        string `json:"networkBandwidthGbps"`
	NetworkBaselineGbps         string `json:"networkBaselineGbps"`
	NetworkPeakGbps             string `json:"networkPeakGbps"`
	EnhancedNetworkingSupported string `json:"enhancedNetworkingSupported"`
	
	// GPU
	GPU              string `json:"gpu"`
	GPUMemory        string `json:"gpuMemory"`
	GPUCount         string `json:"gpuCount"`
	
	// Instance Features
	CurrentGeneration           string `json:"currentGeneration"`
	InstanceFamily             string `json:"instanceFamily"`
	InstanceTypeFamily         string `json:"instanceTypeFamily"`
	DedicatedEbsThroughput     string `json:"dedicatedEbsThroughput"`
	EbsOptimizedSupport        string `json:"ebsOptimizedSupport"`
	
	// Volume Attributes
	VolumeType           string `json:"volumeType"`
	VolumeApiName        string `json:"volumeApiName"`
	MaxIopsvolume        string `json:"maxIopsvolume"`
	MaxThroughputvolume  string `json:"maxThroughputvolume"`
	
	// Additional Features
	Hibernation          string `json:"hibernationSupported"`
	BurstablePerformance string `json:"burstablePerformance"`
	AutoRecovery         string `json:"autoRecovery"`
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

type PriceFilter struct {
	Attribute string
	Value    string
}

type PricingOptions struct {
    OperatingSystem string
    Tenancy        string
    LicenseModel   string
    CapacityType   string   // On-Demand, Reserved, Spot
    ReservedTerm   string   // 1yr, 3yr
    PaymentOption  string   // No Upfront, Partial Upfront, All Upfront
    OfferingClass  string   // Standard, Convertible
    PreInstalledSw string
}

type PricingDetails struct {
    OnDemandPrice   float64
    SpotPrice       float64
    ReservedPricing map[string]ReservedPricing  // Key: term-payment-class
    Attributes      ProductAttributes
}

type ReservedPricing struct {
    UpfrontFee     float64
    HourlyPrice    float64
    EffectivePrice float64  // Calculated for the term
    Term           string
    PaymentOption  string
    OfferingClass  string
}

type EC2Instance struct {
    Type           string
    Region         string
    PricingOptions PricingOptions
    Usage          InstanceUsage
}

type InstanceUsage struct {
    AverageUtilization float64
    PeakUtilization   float64
    BurstableCredits  float64
    StorageGB         int
    IOPS             int
    Throughput       int
}

type SavingsAnalysis struct {
    CurrentInstance    EC2Instance
    TargetInstance    EC2Instance
    HourlySavings     float64
    DailySavings      float64
    MonthlySavings    float64
    YearlySavings     float64
    ReservedSavings   *ReservedSavings
    SpotSavings       *SpotSavings
    Recommendations   []string
}

type ReservedSavings struct {
    Term                  string
    UpfrontSavings       float64
    EffectiveHourlySavings float64
    PaybackPeriodMonths   float64
}

type SpotSavings struct {
    AverageHourlySavings float64
    InterruptionRisk     float64
    RecommendedStrategy  string
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
func (s *EC2PricingService) GetInstancePrice(instanceType, region string, filters ...PriceFilter) (float64, error) {
    // Ensure the region is supported
    if !s.IsRegionSupported(region) {
        return 0, fmt.Errorf("region %s is not supported for EC2 pricing", region)
    }

    // Get pricing data for the specific region
    data, err := s.client.GetServicePricing(EC2Service, region)
    if err != nil {
        return 0, fmt.Errorf("failed to get EC2 pricing data: %w", err)
    }

    // Parse the pricing data
    var pricing struct {
        Products map[string]struct {
            Attributes ProductAttributes `json:"attributes"`
            Sku       string            `json:"sku"`
        } `json:"products"`
        Terms struct {
            OnDemand map[string]map[string]struct {
                PriceDimensions map[string]PriceDimension `json:"priceDimensions"`
                TermAttributes TermAttributes `json:"termAttributes"`
            } `json:"OnDemand"`
            Reserved map[string]map[string]struct {
                PriceDimensions map[string]PriceDimension `json:"priceDimensions"`
                TermAttributes TermAttributes `json:"termAttributes"`
            } `json:"Reserved"`
        } `json:"terms"`
    }

    if err := json.Unmarshal(data, &pricing); err != nil {
        return 0, fmt.Errorf("failed to parse pricing data: %w", err)
    }

    // Default filters if none provided
    if len(filters) == 0 {
        filters = []PriceFilter{
            {
                Attribute: "operatingSystem",
                Value:    OSLinux,
            },
            {
                Attribute: "preInstalledSw",
                Value:    "NA",
            },
            {
                Attribute: "capacitystatus",
                Value:    CapacityUsed,
            },
            {
                Attribute: "tenancy",
                Value:    TenancyShared,
            },
            {
                Attribute: "licenseModel",
                Value:    LicenseNoLicense,
            },
        }
    }

    // Find the matching instance type with all filters
    var matchingSku string
    for sku, product := range pricing.Products {
        attrs := product.Attributes
        if attrs.InstanceType != instanceType {
            continue
        }

        // Apply all filters
        matches := true
        for _, filter := range filters {
            attrValue := getAttributeValue(attrs, filter.Attribute)
            if attrValue != filter.Value {
                matches = false
                break
            }
        }

        if matches {
            matchingSku = sku
            break
        }
    }

    if matchingSku == "" {
        return 0, fmt.Errorf("no matching product found for instance type %s in region %s with specified filters", instanceType, region)
    }

    // Find the price in terms
    for _, term := range pricing.Terms.OnDemand {
        for _, price := range term {
            for _, dimension := range price.PriceDimensions {
                if dimension.Unit == "Hrs" {
                    return parsePrice(dimension.PricePerUnit["USD"])
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

func DefaultPricingOptions() PricingOptions {
    return PricingOptions{
        OperatingSystem: OSLinux,
        Tenancy:        TenancyShared,
        LicenseModel:   LicenseNoLicense,
        CapacityType:   PurchaseOnDemand,
        PreInstalledSw: "NA",
    }
}

func (s *EC2PricingService) GetInstancePriceDetailed(instanceType, region string, options PricingOptions) (*PricingDetails, error) {
    filters := []PriceFilter{
        {Attribute: "instanceType", Value: instanceType},
        {Attribute: "operatingSystem", Value: options.OperatingSystem},
        {Attribute: "tenancy", Value: options.Tenancy},
        {Attribute: "licenseModel", Value: options.LicenseModel},
        {Attribute: "preInstalledSw", Value: options.PreInstalledSw},
    }

    data, err := s.client.GetServicePricing(EC2Service, region)
    if err != nil {
        return nil, fmt.Errorf("failed to get EC2 pricing data: %w", err)
    }

    var pricing struct {
        Products map[string]struct {
            Attributes ProductAttributes `json:"attributes"`
            Sku       string            `json:"sku"`
        } `json:"products"`
        Terms struct {
            OnDemand map[string]map[string]struct {
                PriceDimensions map[string]PriceDimension `json:"priceDimensions"`
                TermAttributes TermAttributes             `json:"termAttributes"`
            } `json:"OnDemand"`
            Reserved map[string]map[string]struct {
                PriceDimensions map[string]PriceDimension `json:"priceDimensions"`
                TermAttributes TermAttributes             `json:"termAttributes"`
            } `json:"Reserved"`
        } `json:"terms"`
    }

    if err := json.Unmarshal(data, &pricing); err != nil {
        return nil, fmt.Errorf("failed to parse pricing data: %w", err)
    }

    details := &PricingDetails{
        ReservedPricing: make(map[string]ReservedPricing),
    }

    // Find matching product
    var matchingSku string
    for sku, product := range pricing.Products {
        attrs := product.Attributes
        matches := true
        for _, filter := range filters {
            if getAttributeValue(attrs, filter.Attribute) != filter.Value {
                matches = false
                break
            }
        }
        if matches {
            matchingSku = sku
            details.Attributes = attrs
            break
        }
    }

    if matchingSku == "" {
        return nil, fmt.Errorf("no matching product found for instance type %s in region %s", instanceType, region)
    }

    // Get On-Demand price
    for _, term := range pricing.Terms.OnDemand {
        for _, price := range term {
            for _, dimension := range price.PriceDimensions {
                if dimension.Unit == "Hrs" {
                    details.OnDemandPrice, _ = parsePrice(dimension.PricePerUnit["USD"])
                    break
                }
            }
        }
    }

    // Get Reserved Instance prices
    for _, term := range pricing.Terms.Reserved {
        for _, price := range term {
            attrs := price.TermAttributes
            key := fmt.Sprintf("%s-%s-%s", attrs.LeaseContractLength, attrs.PurchaseOption, attrs.OfferingClass)
            
            rp := ReservedPricing{
                Term:          attrs.LeaseContractLength,
                PaymentOption: attrs.PurchaseOption,
                OfferingClass: attrs.OfferingClass,
            }

            for _, dimension := range price.PriceDimensions {
                if dimension.Unit == "Hrs" {
                    rp.HourlyPrice, _ = parsePrice(dimension.PricePerUnit["USD"])
                } else if dimension.Unit == "Quantity" {
                    rp.UpfrontFee, _ = parsePrice(dimension.PricePerUnit["USD"])
                }
            }

            // Calculate effective hourly price
            hours := 8760.0 // 1 year
            if attrs.LeaseContractLength == "3yr" {
                hours = 26280.0
            }
            rp.EffectivePrice = rp.HourlyPrice + (rp.UpfrontFee / hours)

            details.ReservedPricing[key] = rp
        }
    }

    return details, nil
}

func (s *EC2PricingService) CalculateDetailedSavings(current, target EC2Instance) (*SavingsAnalysis, error) {
    currentPricing, err := s.GetInstancePriceDetailed(current.Type, current.Region, current.PricingOptions)
    if err != nil {
        return nil, fmt.Errorf("failed to get current instance pricing: %w", err)
    }

    targetPricing, err := s.GetInstancePriceDetailed(target.Type, target.Region, target.PricingOptions)
    if err != nil {
        return nil, fmt.Errorf("failed to get target instance pricing: %w", err)
    }

    analysis := &SavingsAnalysis{
        CurrentInstance: current,
        TargetInstance: target,
        HourlySavings: currentPricing.OnDemandPrice - targetPricing.OnDemandPrice,
    }

    // Calculate time-based savings
    analysis.DailySavings = analysis.HourlySavings * 24
    analysis.MonthlySavings = analysis.HourlySavings * 730  // Average hours per month
    analysis.YearlySavings = analysis.HourlySavings * 8760

    // Calculate Reserved Instance savings
    if riPricing, ok := targetPricing.ReservedPricing["1yr-partial-standard"]; ok {
        yearlyOnDemand := currentPricing.OnDemandPrice * 8760
        yearlyReserved := riPricing.UpfrontFee + (riPricing.HourlyPrice * 8760)
        
        analysis.ReservedSavings = &ReservedSavings{
            Term: "1yr",
            UpfrontSavings: yearlyOnDemand - yearlyReserved,
            EffectiveHourlySavings: currentPricing.OnDemandPrice - riPricing.EffectivePrice,
            PaybackPeriodMonths: (riPricing.UpfrontFee / analysis.MonthlySavings),
        }
    }

    // Add recommendations based on usage patterns
    analysis.Recommendations = s.generateRecommendations(current, target, currentPricing, targetPricing)

    return analysis, nil
}

func (s *EC2PricingService) generateRecommendations(current, target EC2Instance, currentPricing, targetPricing *PricingDetails) []string {
    var recommendations []string

    // Check for cost-effective instance type
    if current.Usage.AverageUtilization < 40 {
        recommendations = append(recommendations, fmt.Sprintf(
            "Consider downsizing from %s to %s due to low utilization (%.1f%%)",
            current.Type, target.Type, current.Usage.AverageUtilization,
        ))
    }

    // Check for Reserved Instance opportunities
    if current.PricingOptions.CapacityType == PurchaseOnDemand {
        if riPricing, ok := targetPricing.ReservedPricing["1yr-partial-standard"]; ok {
            savings := (currentPricing.OnDemandPrice - riPricing.EffectivePrice) * 8760
            if savings > 1000 { // If yearly savings exceed $1000
                recommendations = append(recommendations, fmt.Sprintf(
                    "Consider Reserved Instance for %s to save approximately $%.2f per year",
                    target.Type, savings,
                ))
            }
        }
    }

    // Check for burstable instance opportunities
    if current.Usage.PeakUtilization < 20 {
        recommendations = append(recommendations, fmt.Sprintf(
            "Consider using a burstable instance type for %s due to very low peak utilization (%.1f%%)",
            current.Type, current.Usage.PeakUtilization,
        ))
    }

    return recommendations
}

func parsePrice(price string) (float64, error) {
	var value float64
	if _, err := fmt.Sscanf(price, "%f", &value); err != nil {
		return 0, fmt.Errorf("failed to parse price %s: %v", price, err)
	}
	return value, nil
}

func getAttributeValue(attrs ProductAttributes, attributeName string) string {
	r := reflect.ValueOf(attrs)
	f := reflect.Indirect(r).FieldByNameFunc(func(s string) bool {
		return strings.EqualFold(s, attributeName)
	})
	
	if !f.IsValid() {
		return ""
	}
	
	return f.String()
}
