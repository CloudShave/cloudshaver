package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// EC2ClientAPI defines the interface for EC2 client operations
type EC2ClientAPI interface {
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	DescribeVolumes(ctx context.Context, params *ec2.DescribeVolumesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVolumesOutput, error)
}

// PricingServiceAPI defines the interface for pricing operations
type PricingServiceAPI interface {
	IsRegionSupported(region string) bool
	GetVolumePrice(volumeType, region string) (float64, error)
	CalculateInstanceSavings(currentType, targetType, region string) (float64, error)
}
