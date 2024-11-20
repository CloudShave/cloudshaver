package factory

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awsblades "github.com/cloudshave/cloudshaver/internal/blades/aws"
	awspricing "github.com/cloudshave/cloudshaver/internal/blades/aws/pricing"
	"github.com/cloudshave/cloudshaver/internal/types"
)

// BladeConfig represents the configuration for creating a blade
type BladeConfig struct {
	Provider types.CloudProvider
	Region   string
	// Add more configuration options as needed
}

// CreateBlade creates a blade instance based on the provided configuration
func CreateBlade(ctx context.Context, bladeConfig BladeConfig) (types.Blade, error) {
	switch bladeConfig.Provider {
	case types.AWS:
		return createAWSBlade(ctx, bladeConfig)
	case types.Azure:
		return createAzureBlade(ctx, bladeConfig)
	case types.GCP:
		return createGCPBlade(ctx, bladeConfig)
	default:
		return nil, fmt.Errorf("unsupported cloud provider: %s", bladeConfig.Provider)
	}
}

func createAWSBlade(ctx context.Context, bladeConfig BladeConfig) (types.Blade, error) {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(bladeConfig.Region))
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	// Create EC2 client
	ec2Client := ec2.NewFromConfig(cfg)

	// Create pricing service
	pricingService, err := awspricing.NewEC2PricingService(bladeConfig.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to create pricing service: %w", err)
	}

	// Create EC2 blade
	blade, err := awsblades.NewEC2Blade(ec2Client, pricingService, bladeConfig.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to create EC2 blade: %w", err)
	}

	return blade, nil
}

func createAzureBlade(ctx context.Context, config BladeConfig) (types.Blade, error) {
	// TODO: Implement Azure blade creation
	return nil, fmt.Errorf("azure blade creation not implemented")
}

func createGCPBlade(ctx context.Context, config BladeConfig) (types.Blade, error) {
	// TODO: Implement GCP blade creation
	return nil, fmt.Errorf("gcp blade creation not implemented")
}
