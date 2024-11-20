package factory

import (
	"context"

	"github.com/cloudshave/cloudshaver/internal/aws/client"
	awsblades "github.com/cloudshave/cloudshaver/internal/blades/aws"
	awspricing "github.com/cloudshave/cloudshaver/internal/pricing/aws"
	"github.com/cloudshave/cloudshaver/internal/types"
)

// CreateBlades creates all available cost optimization blades
func CreateBlades(ctx context.Context, region string) ([]types.Blade, error) {
	var blades []types.Blade

	// Create EC2 client
	ec2Client, err := awsclient.NewEC2Client(ctx, region)
	if err != nil {
		return nil, err
	}

	// Create pricing service
	pricingService, err := awspricing.NewPricingService()
	if err != nil {
		return nil, err
	}

	// Create EC2 blade
	ec2Blade, err := awsblades.NewEC2Blade(ec2Client, pricingService, region)
	if err != nil {
		return nil, err
	}
	blades = append(blades, ec2Blade)

	return blades, nil
}
