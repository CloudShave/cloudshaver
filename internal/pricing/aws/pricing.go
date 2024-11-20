package aws

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "time"
)

type EC2Pricing struct {
    LastUpdated         string                           `json:"last_updated"`
    RegionMapping       map[string]string                `json:"region_mapping"`
    OnDemandInstances   map[string]map[string]Instance   `json:"on_demand_instances"`
    EBSVolumes         map[string]map[string]Volume     `json:"ebs_volumes"`
    SavingsOpportunities SavingsOpportunities            `json:"savings_opportunities"`
}

type Instance struct {
    VCPU               int     `json:"vcpu"`
    MemoryGiB         int     `json:"memory_gib"`
    PricePerHour      float64 `json:"price_per_hour"`
    RecommendedUpgrade string  `json:"recommended_upgrade,omitempty"`
    RecommendedDowngrade string `json:"recommended_downgrade,omitempty"`
}

type Volume struct {
    PricePerGBMonth     float64 `json:"price_per_gb_month"`
    BasePricePerMonth   float64 `json:"base_price_per_month,omitempty"`
    IOPSIncluded        int     `json:"iops_included,omitempty"`
    ThroughputIncluded  int     `json:"throughput_included_mibps,omitempty"`
    PricePerIOPSMonth   float64 `json:"price_per_iops_month,omitempty"`
    RecommendedUpgrade  string  `json:"recommended_upgrade,omitempty"`
}

type SavingsOpportunities struct {
    InstanceUpgrade struct {
        T2ToT3 struct {
            AverageSavingsPercentage float64 `json:"average_savings_percentage"`
            Description             string  `json:"description"`
        } `json:"t2_to_t3"`
        Oversized struct {
            CPUThresholdPercent    int     `json:"cpu_threshold_percent"`
            MemoryThresholdPercent int     `json:"memory_threshold_percent"`
            MinimumDays           int     `json:"minimum_days"`
            Description           string  `json:"description"`
        } `json:"oversized"`
    } `json:"instance_upgrade"`
    VolumeOptimization struct {
        GP2ToGP3 struct {
            MinimumSizeGB  int     `json:"minimum_size_gb"`
            SavingsPerGB   float64 `json:"savings_per_gb"`
            Description    string  `json:"description"`
        } `json:"gp2_to_gp3"`
        Unattached struct {
            MinimumDays  int    `json:"minimum_days"`
            Description string `json:"description"`
        } `json:"unattached"`
        Underutilized struct {
            IOPSThreshold   int    `json:"iops_threshold"`
            SizeThresholdGB int    `json:"size_threshold_gb"`
            Description     string `json:"description"`
        } `json:"underutilized"`
    } `json:"volume_optimization"`
}

var pricingData *EC2Pricing

// LoadPricing loads the pricing data from the JSON file
func LoadPricing() (*EC2Pricing, error) {
    if pricingData != nil {
        return pricingData, nil
    }

    // Get the directory of the current file
    dir, err := os.Getwd()
    if err != nil {
        return nil, fmt.Errorf("failed to get current directory: %v", err)
    }

    // Construct path to the pricing data file
    pricingFile := filepath.Join(dir, "internal", "pricing", "aws", "data", "ec2_pricing.json")
    data, err := os.ReadFile(pricingFile)
    if err != nil {
        return nil, fmt.Errorf("failed to read pricing data: %v", err)
    }

    pricing := &EC2Pricing{}
    if err := json.Unmarshal(data, pricing); err != nil {
        return nil, fmt.Errorf("failed to parse pricing data: %v", err)
    }

    // Validate last updated date
    lastUpdated, err := time.Parse("2006-01-02", pricing.LastUpdated)
    if err != nil {
        return nil, fmt.Errorf("invalid last_updated date format: %v", err)
    }

    // Warn if pricing data is older than 30 days
    if time.Since(lastUpdated) > 30*24*time.Hour {
        fmt.Printf("Warning: Pricing data is more than 30 days old (last updated: %s)\n", pricing.LastUpdated)
    }

    pricingData = pricing
    return pricing, nil
}

// CalculateInstanceSavings calculates potential savings for an EC2 instance
func (p *EC2Pricing) CalculateInstanceSavings(region, instanceType string, hoursRunning int) (float64, string, error) {
    regionPricing, ok := p.OnDemandInstances[region]
    if !ok {
        return 0, "", fmt.Errorf("pricing not available for region: %s", region)
    }

    instance, ok := regionPricing[instanceType]
    if !ok {
        return 0, "", fmt.Errorf("pricing not available for instance type: %s", instanceType)
    }

    var savings float64
    var recommendation string

    // Check for T2 to T3 upgrade opportunity
    if upgrade := instance.RecommendedUpgrade; upgrade != "" {
        if upgradeInstance, ok := regionPricing[upgrade]; ok {
            hourlyDiff := instance.PricePerHour - upgradeInstance.PricePerHour
            savings = hourlyDiff * float64(hoursRunning)
            recommendation = fmt.Sprintf("Upgrade to %s to save $%.2f per month", upgrade, savings*730) // 730 hours in a month
        }
    }

    // Check for downsizing opportunity
    if downgrade := instance.RecommendedDowngrade; downgrade != "" {
        if downgradeInstance, ok := regionPricing[downgrade]; ok {
            hourlyDiff := instance.PricePerHour - downgradeInstance.PricePerHour
            downgradeSavings := hourlyDiff * float64(hoursRunning)
            if downgradeSavings > savings {
                savings = downgradeSavings
                recommendation = fmt.Sprintf("Downgrade to %s to save $%.2f per month", downgrade, savings*730)
            }
        }
    }

    return savings, recommendation, nil
}

// CalculateVolumeSavings calculates potential savings for an EBS volume
func (p *EC2Pricing) CalculateVolumeSavings(region string, volumeType string, sizeGB int, iops int) (float64, string, error) {
    regionPricing, ok := p.EBSVolumes[region]
    if !ok {
        return 0, "", fmt.Errorf("pricing not available for region: %s", region)
    }

    volume, ok := regionPricing[volumeType]
    if !ok {
        return 0, "", fmt.Errorf("pricing not available for volume type: %s", volumeType)
    }

    var monthlySavings float64
    var recommendation string

    // Check for GP2 to GP3 migration opportunity
    if volumeType == "gp2" && sizeGB >= p.SavingsOpportunities.VolumeOptimization.GP2ToGP3.MinimumSizeGB {
        gp3Pricing := regionPricing["gp3"]
        savingsPerGB := volume.PricePerGBMonth - gp3Pricing.PricePerGBMonth
        monthlySavings = float64(sizeGB) * savingsPerGB
        recommendation = fmt.Sprintf("Migrate to gp3 volume type to save $%.2f per month", monthlySavings)
    }

    // Check for over-provisioned IOPS
    if (volumeType == "io1" || volumeType == "io2") && iops < p.SavingsOpportunities.VolumeOptimization.Underutilized.IOPSThreshold {
        iopsPrice := volume.PricePerIOPSMonth * float64(iops)
        recommendation = fmt.Sprintf("Consider reducing provisioned IOPS to save up to $%.2f per month", iopsPrice)
        monthlySavings = iopsPrice
    }

    return monthlySavings, recommendation, nil
}
