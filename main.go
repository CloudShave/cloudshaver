package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/cloudshave/cloudshaver/internal/aws/client"
	"github.com/cloudshave/cloudshaver/internal/factory"
	"github.com/cloudshave/cloudshaver/internal/types"
	"github.com/sirupsen/logrus"
)

func main() {
	// Configure logging
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)

	// Validate AWS credentials before proceeding
	ctx := context.Background()
	if err := awsclient.ValidateCredentials(ctx); err != nil {
		logrus.Fatalf("AWS Credentials validation failed: %v\nPlease ensure valid AWS credentials are configured either through:\n"+
			"1. AWS CLI credentials file (~/.aws/credentials)\n"+
			"2. Environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)\n"+
			"3. IAM Role (if running on AWS infrastructure)", err)
		os.Exit(1)
	}

	// Default to us-east-1 if no region specified
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	// Create blade configuration
	bladeConfig := factory.BladeConfig{
		Provider: types.AWS,
		Region:   region,
	}

	// Create blades
	blades, err := factory.CreateBlade(ctx, bladeConfig)
	if err != nil {
		logrus.Fatalf("Failed to create blades: %v", err)
	}

	// Execute all blades and collect results
	var results []*types.BladeResult
	for _, blade := range blades {
		logrus.Infof("Executing blade: %s", blade.GetName())
		result, err := blade.Execute()
		if err != nil {
			logrus.Errorf("Failed to execute blade %s: %v", blade.GetName(), err)
			continue
		}
		results = append(results, result)
	}

	// Output results
	outputJSON(results)
}

func summarizeResults(results []*types.BladeResult) {
	var totalSavings float64
	for _, result := range results {
		totalSavings += result.PotentialSavings
		fmt.Printf("\nBlade: %s\n", result.BladeName)
		fmt.Printf("Category: %s\n", result.Category)
		fmt.Printf("Potential Savings: $%.2f\n", result.PotentialSavings)
		fmt.Println("\nDetails:")
		for _, detail := range result.Details {
			fmt.Printf("- %s\n", detail)
		}
	}
	fmt.Printf("\nTotal Potential Savings: $%.2f\n", totalSavings)
}

func outputJSON(results []*types.BladeResult) {
	output := struct {
		Timestamp time.Time           `json:"timestamp"`
		Results   []*types.BladeResult `json:"results"`
	}{
		Timestamp: time.Now(),
		Results:   results,
	}

	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		logrus.Fatalf("Failed to marshal results to JSON: %v", err)
	}

	fmt.Println(string(jsonData))
}
