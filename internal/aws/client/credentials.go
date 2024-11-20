package awsclient

import (
    "context"
    "fmt"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/sts"
)

// ValidateCredentials checks if AWS credentials are valid by making a test API call
func ValidateCredentials(ctx context.Context) error {
    cfg, err := config.LoadDefaultConfig(ctx)
    if err != nil {
        return fmt.Errorf("unable to load AWS SDK config: %v", err)
    }

    stsClient := sts.NewFromConfig(cfg)
    _, err = stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
    if err != nil {
        return fmt.Errorf("invalid AWS credentials: %v", err)
    }
    return nil
}
