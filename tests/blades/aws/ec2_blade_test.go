package aws_test

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	awsblades "github.com/cloudshave/cloudshaver/internal/blades/aws"
	"github.com/cloudshave/cloudshaver/tests/utils"
)

func TestEC2Blade_Execute(t *testing.T) {
	mockEC2Client := utils.NewMockEC2Client(t)
	mockPricingService := utils.NewMockPricingService(t)

	// Add test instances
	mockEC2Client.Instances = append(mockEC2Client.Instances,
		utils.CreateTestEC2Instance("i-1", types.InstanceTypeT2Micro),
		utils.CreateTestEC2Instance("i-2", types.InstanceTypeT2Small))

	blade, err := awsblades.NewEC2Blade(mockEC2Client, mockPricingService, "us-west-2")
	if err != nil {
		t.Fatalf("Failed to create EC2 blade: %v", err)
	}

	result, err := blade.Execute()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Error("Expected result to not be nil")
	}

	// Test error case
	mockEC2Client.Err = errors.New("test error")
	result, err = blade.Execute()

	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestNewEC2Blade(t *testing.T) {
	mockEC2Client := utils.NewMockEC2Client(t)
	mockPricingService := utils.NewMockPricingService(t)

	blade, err := awsblades.NewEC2Blade(mockEC2Client, mockPricingService, "us-west-2")
	if err != nil {
		t.Fatalf("Failed to create EC2 blade: %v", err)
	}

	if blade == nil {
		t.Error("Expected blade to not be nil")
	}
}
