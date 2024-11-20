package factory

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	awsblades "github.com/cloudshave/cloudshaver/internal/blades/aws"
	awspricing "github.com/cloudshave/cloudshaver/internal/pricing/aws"
	"github.com/cloudshave/cloudshaver/internal/types"
)

// BladeConfig represents the configuration for creating a blade
type BladeConfig struct {
	Provider types.CloudProvider
	Region   string
	// Add more configuration options as needed
}

// CreateBlade creates blade instances based on the provided configuration
func CreateBlade(ctx context.Context, bladeConfig BladeConfig) ([]types.Blade, error) {
	switch bladeConfig.Provider {
	case types.AWS:
		return createAWSBlade(ctx, bladeConfig)
	// case types.Azure:
	// 	return createAzureBlade(ctx, bladeConfig)
	// case types.GCP:
	// 	return createGCPBlade(ctx, bladeConfig)
	default:
		return nil, fmt.Errorf("unsupported cloud provider: %s", bladeConfig.Provider)
	}
}

func createAWSBlade(ctx context.Context, bladeConfig BladeConfig) ([]types.Blade, error) {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(bladeConfig.Region))
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	// Create EC2 client
	ec2Client := ec2.NewFromConfig(cfg)

	// Create RDS client
	rdsClient := rds.NewFromConfig(cfg)

	// Create CloudWatch client
	cloudWatchClient := cloudwatch.NewFromConfig(cfg)

	// Create pricing service
	pricingService, err := awspricing.NewPricingService()
	if err != nil {
		return nil, fmt.Errorf("failed to create pricing service: %w", err)
	}

	// Create EC2 blade
	ec2Blade, err := awsblades.NewEC2Blade(ec2Client, pricingService, bladeConfig.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to create EC2 blade: %w", err)
	}

	// Create RDS blade
	rdsBlade, err := awsblades.NewRDSBlade(rdsClient, cloudWatchClient, pricingService, bladeConfig.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to create RDS blade: %w", err)
	}

	// Return the requested blades
	return []types.Blade{ec2Blade, rdsBlade}, nil
}

// func createAzureBlade(ctx context.Context, bladeConfig BladeConfig) ([]types.Blade, error) {
// 	// TODO: Implement Azure blade creation
// 	return nil, fmt.Errorf("Azure blades not yet implemented")
// }

// func createGCPBlade(ctx context.Context, bladeConfig BladeConfig) ([]types.Blade, error) {
// 	// TODO: Implement GCP blade creation
// 	return nil, fmt.Errorf("GCP blades not yet implemented")
// }
