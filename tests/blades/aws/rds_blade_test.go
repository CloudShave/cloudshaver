package aws_test

import (
	"errors"
	"testing"

	awsblades "github.com/cloudshave/cloudshaver/internal/blades/aws"
	"github.com/cloudshave/cloudshaver/tests/utils"
)

func TestRDSBlade_Execute(t *testing.T) {
	mockRDSClient := utils.NewMockRDSClient(t)
	mockCloudWatchClient := utils.NewMockCloudWatchClient(t)
	mockPricingService := utils.NewMockPricingService(t)

	// Add test instances
	mockRDSClient.Instances = append(mockRDSClient.Instances,
		utils.CreateTestRDSInstance("db-1", "db.t3.micro"),
		utils.CreateTestRDSInstance("db-2", "db.t3.small"))

	// Add test snapshots
	mockRDSClient.Snapshots = append(mockRDSClient.Snapshots,
		utils.CreateTestDBSnapshot("snap-1", "db-1"),
		utils.CreateTestDBSnapshot("snap-2", "db-2"))

	blade, err := awsblades.NewRDSBlade(mockRDSClient, mockCloudWatchClient, mockPricingService, "us-west-2")
	if err != nil {
		t.Fatalf("Failed to create RDS blade: %v", err)
	}

	result, err := blade.Execute()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Error("Expected result to not be nil")
	}

	// Test error case
	mockRDSClient.Err = errors.New("test error")
	result, err = blade.Execute()

	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestNewRDSBlade(t *testing.T) {
	mockRDSClient := utils.NewMockRDSClient(t)
	mockCloudWatchClient := utils.NewMockCloudWatchClient(t)
	mockPricingService := utils.NewMockPricingService(t)

	blade, err := awsblades.NewRDSBlade(mockRDSClient, mockCloudWatchClient, mockPricingService, "us-west-2")
	if err != nil {
		t.Fatalf("Failed to create RDS blade: %v", err)
	}

	if blade == nil {
		t.Error("Expected blade to not be nil")
	}
}
