package awsclient

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awsinterfaces "github.com/cloudshave/cloudshaver/internal/aws/interfaces"
)

// NewEC2Client creates a new EC2 client that implements EC2ClientAPI
func NewEC2Client(ctx context.Context, region string) (awsinterfaces.EC2ClientAPI, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}
	return ec2.NewFromConfig(cfg), nil
}
