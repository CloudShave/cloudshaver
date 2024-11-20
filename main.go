package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/cloudshave/cloudshaver/internal/aws"
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
	if err := aws.ValidateCredentials(ctx); err != nil {
		logrus.Fatalf("AWS Credentials validation failed: %v\nPlease ensure valid AWS credentials are configured either through:\n"+
			"1. AWS CLI credentials file (~/.aws/credentials)\n"+
			"2. Environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)\n"+
			"3. IAM Role (if running on AWS infrastructure)", err)
		os.Exit(1)
	}

	// Define blade configurations
	bladeConfigs := []factory.BladeConfig{
		{
			Provider: types.AWS,
			Region:   "eu-west-1", // Changed to eu-west-1 for your EC2 instances
		},
		// Add more blade configurations here as needed
		// Example:
		// {
		//     Provider: types.Azure,
		//     Region:   "eastus",
		// },
	}

	// Initialize blades
	var blades []types.Blade
	for _, config := range bladeConfigs {
		blade, err := factory.CreateBlade(ctx, config)
		if err != nil {
			logrus.WithError(err).Errorf("Failed to create blade for provider %s", config.Provider)
			continue
		}
		blades = append(blades, blade)
	}

	// Execute blades and collect results
	var allResults []*types.BladeResult

	for _, blade := range blades {
		logrus.Infof("Executing blade: %s", blade.GetName())

		result, err := blade.Execute()
		if err != nil {
			logrus.WithError(err).Errorf("Blade %s failed", blade.GetName())
			continue
		}

		allResults = append(allResults, result)
	}

	// Output results
	if len(allResults) > 0 {
		summarizeResults(allResults)
		outputJSON(allResults)
	} else {
		logrus.Info("No results were generated from any blades")
	}
}

func summarizeResults(results []*types.BladeResult) {
	totalSavings := 0.0

	// Pretty print results
	fmt.Println("\n=== CloudShaver Cost Optimization Report ===")

	for _, result := range results {
		totalSavings += result.PotentialSavings

		fmt.Printf("\nBlade: %s\n", result.Category)
		fmt.Printf("Cloud Provider: %s\n", result.CloudProvider)
		fmt.Printf("Potential Savings: $%.2f\n", result.PotentialSavings)

		fmt.Println("Recommendations:")
		for _, rec := range result.Recommendations {
			fmt.Printf("- %s\n", rec)
		}
	}

	fmt.Printf("\nTotal Potential Savings: $%.2f\n", totalSavings)
}

func outputJSON(results []*types.BladeResult) {
	filename := fmt.Sprintf("cloudshaver_report_%s.json", time.Now().Format("20060102_150405"))

	file, err := json.MarshalIndent(results, "", " ")
	if err != nil {
		logrus.WithError(err).Error("Failed to create JSON report")
		return
	}

	err = os.WriteFile(filename, file, 0644)
	if err != nil {
		logrus.WithError(err).Error("Failed to write JSON report")
		return
	}

	logrus.Infof("Report saved to %s", filename)
}
