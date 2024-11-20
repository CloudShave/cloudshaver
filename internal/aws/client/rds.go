package awsclient

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	awsinterfaces "github.com/cloudshave/cloudshaver/internal/interfaces/aws"
)

// NewRDSClient creates a new RDS client that implements RDSClientAPI
func NewRDSClient(ctx context.Context, region string) (awsinterfaces.RDSClientAPI, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}
	return rds.NewFromConfig(cfg), nil
}
