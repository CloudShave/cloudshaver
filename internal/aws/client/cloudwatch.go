package awsclient

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	awsinterfaces "github.com/cloudshave/cloudshaver/internal/interfaces/aws"
)

// NewCloudWatchClient creates a new CloudWatch client that implements CloudWatchClientAPI
func NewCloudWatchClient(ctx context.Context, region string) (awsinterfaces.CloudWatchClientAPI, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}
	return cloudwatch.NewFromConfig(cfg), nil
}
