package factory_test

import (
	"context"
	"testing"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/cloudshave/cloudshaver/internal/factory"
	"github.com/cloudshave/cloudshaver/tests/utils"
)

const (
	testRegion = "us-west-2"
)

func TestCreateBlade(t *testing.T) {
	ctx := context.Background()
	bladeConfig := factory.BladeConfig{
		Provider: "aws", // Using string literal since types.AWS is a constant
		Region:   testRegion,
	}

	// Create mock clients
	mockEC2Client := utils.NewMockEC2Client(t)
	mockRDSClient := utils.NewMockRDSClient(t)
	mockCloudWatchClient := utils.NewMockCloudWatchClient(t)
	mockPricingService := utils.NewMockPricingService(t)

	// Configure mock EC2 client
	mockEC2Client.Instances = append(mockEC2Client.Instances,
		utils.CreateTestEC2Instance("i-1", ec2types.InstanceTypeT2Micro),
		utils.CreateTestEC2Instance("i-2", ec2types.InstanceTypeT2Small))

	mockEC2Client.Volumes = append(mockEC2Client.Volumes,
		utils.CreateTestVolume("vol-1", ec2types.VolumeTypeGp2),
		utils.CreateTestVolume("vol-2", ec2types.VolumeTypeGp2))

	// Configure mock RDS client
	mockRDSClient.Instances = append(mockRDSClient.Instances,
		utils.CreateTestRDSInstance("db-1", "db.t3.micro"))
	mockRDSClient.Snapshots = append(mockRDSClient.Snapshots,
		utils.CreateTestDBSnapshot("snap-1", "db-1"))

	// Create AWS clients for testing
	awsClients := factory.AWSClients{
		EC2Client:        mockEC2Client,
		RDSClient:        mockRDSClient,
		CloudWatchClient: mockCloudWatchClient,
		PricingService:   mockPricingService,
	}

	blades, err := factory.CreateBlade(ctx, bladeConfig, awsClients)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(blades) != 2 { // We expect EC2 and RDS blades
		t.Errorf("Expected 2 blades, got %d", len(blades))
	}
}

func TestCreateBlade_UnsupportedProvider(t *testing.T) {
	ctx := context.Background()
	bladeConfig := factory.BladeConfig{
		Provider: "unsupported",
		Region:   testRegion,
	}

	blades, err := factory.CreateBlade(ctx, bladeConfig)
	if err == nil {
		t.Error("Expected error for unsupported provider, got nil")
	}

	if blades != nil {
		t.Errorf("Expected nil blades for unsupported provider, got %v", blades)
	}
}

func TestCreateAWSBlade(t *testing.T) {
	ctx := context.Background()
	bladeConfig := factory.BladeConfig{
		Provider: "aws",
		Region:   testRegion,
	}

	// Create mock clients
	mockEC2Client := utils.NewMockEC2Client(t)
	mockRDSClient := utils.NewMockRDSClient(t)
	mockCloudWatchClient := utils.NewMockCloudWatchClient(t)
	mockPricingService := utils.NewMockPricingService(t)

	// Configure mock EC2 client
	mockEC2Client.Instances = append(mockEC2Client.Instances,
		utils.CreateTestEC2Instance("i-1", ec2types.InstanceTypeT2Micro),
		utils.CreateTestEC2Instance("i-2", ec2types.InstanceTypeT2Small))

	mockEC2Client.Volumes = append(mockEC2Client.Volumes,
		utils.CreateTestVolume("vol-1", ec2types.VolumeTypeGp2),
		utils.CreateTestVolume("vol-2", ec2types.VolumeTypeGp2))

	// Configure mock RDS client
	mockRDSClient.Instances = append(mockRDSClient.Instances,
		utils.CreateTestRDSInstance("db-1", "db.t3.micro"))
	mockRDSClient.Snapshots = append(mockRDSClient.Snapshots,
		utils.CreateTestDBSnapshot("snap-1", "db-1"))

	// Create AWS clients for testing
	awsClients := factory.AWSClients{
		EC2Client:        mockEC2Client,
		RDSClient:        mockRDSClient,
		CloudWatchClient: mockCloudWatchClient,
		PricingService:   mockPricingService,
	}

	blades, err := factory.CreateBlade(ctx, bladeConfig, awsClients)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(blades) != 2 {
		t.Errorf("Expected 2 blades (EC2 and RDS), got %d", len(blades))
	}
}
