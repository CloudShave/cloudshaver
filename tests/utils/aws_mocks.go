package utils

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	awsinterfaces "github.com/cloudshave/cloudshaver/internal/interfaces/aws"
)

// MockEC2Client mocks the EC2 client for testing
type MockEC2Client struct {
	t         *testing.T
	awsinterfaces.EC2ClientAPI
	Instances []ec2types.Instance
	Volumes   []ec2types.Volume
	Err       error
}

// NewMockEC2Client creates a new mock EC2 client for testing
func NewMockEC2Client(t *testing.T) *MockEC2Client {
	return &MockEC2Client{
		t:         t,
		Instances: make([]ec2types.Instance, 0),
		Volumes:   make([]ec2types.Volume, 0),
	}
}

func (m *MockEC2Client) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	if m.Err != nil {
		return nil, m.Err
	}

	reservations := []ec2types.Reservation{
		{Instances: m.Instances},
	}

	return &ec2.DescribeInstancesOutput{
		Reservations: reservations,
	}, nil
}

func (m *MockEC2Client) DescribeVolumes(ctx context.Context, params *ec2.DescribeVolumesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVolumesOutput, error) {
	if m.Err != nil {
		return nil, m.Err
	}

	return &ec2.DescribeVolumesOutput{
		Volumes: m.Volumes,
	}, nil
}

// MockRDSClient mocks the RDS client for testing
type MockRDSClient struct {
	t         *testing.T
	awsinterfaces.RDSClientAPI
	Instances []rdstypes.DBInstance
	Snapshots []rdstypes.DBSnapshot
	Err       error
}

// NewMockRDSClient creates a new mock RDS client for testing
func NewMockRDSClient(t *testing.T) *MockRDSClient {
	return &MockRDSClient{
		t:         t,
		Instances: make([]rdstypes.DBInstance, 0),
		Snapshots: make([]rdstypes.DBSnapshot, 0),
	}
}

func (m *MockRDSClient) DescribeDBInstances(ctx context.Context, params *rds.DescribeDBInstancesInput, optFns ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error) {
	if m.Err != nil {
		return nil, m.Err
	}

	return &rds.DescribeDBInstancesOutput{
		DBInstances: m.Instances,
	}, nil
}

func (m *MockRDSClient) DescribeDBSnapshots(ctx context.Context, params *rds.DescribeDBSnapshotsInput, optFns ...func(*rds.Options)) (*rds.DescribeDBSnapshotsOutput, error) {
	if m.Err != nil {
		return nil, m.Err
	}

	return &rds.DescribeDBSnapshotsOutput{
		DBSnapshots: m.Snapshots,
	}, nil
}

func (m *MockRDSClient) DescribeReservedDBInstances(ctx context.Context, params *rds.DescribeReservedDBInstancesInput, optFns ...func(*rds.Options)) (*rds.DescribeReservedDBInstancesOutput, error) {
	if m.Err != nil {
		return nil, m.Err
	}

	return &rds.DescribeReservedDBInstancesOutput{
		ReservedDBInstances: []rdstypes.ReservedDBInstance{},
	}, nil
}

// MockCloudWatchClient mocks the CloudWatch client for testing
type MockCloudWatchClient struct {
	t   *testing.T
	awsinterfaces.CloudWatchClientAPI
	Metrics []types.Metric
	Err     error
}

// NewMockCloudWatchClient creates a new mock CloudWatch client for testing
func NewMockCloudWatchClient(t *testing.T) *MockCloudWatchClient {
	return &MockCloudWatchClient{
		t:      t,
		Metrics: make([]types.Metric, 0),
	}
}

func (m *MockCloudWatchClient) GetMetricData(ctx context.Context, params *cloudwatch.GetMetricDataInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.GetMetricDataOutput, error) {
	if m.Err != nil {
		return nil, m.Err
	}

	return &cloudwatch.GetMetricDataOutput{}, nil
}

// MockPricingService mocks the pricing service for testing
type MockPricingService struct {
	t   *testing.T
	awsinterfaces.PricingServiceAPI
	Savings float64
	Err     error
}

// NewMockPricingService creates a new mock pricing service for testing
func NewMockPricingService(t *testing.T) *MockPricingService {
	return &MockPricingService{
		t: t,
	}
}

func (m *MockPricingService) CalculateInstanceSavings(currentType, targetType, region string) (float64, error) {
	if m.Err != nil {
		return 0, m.Err
	}
	return m.Savings, nil
}

func (m *MockPricingService) LoadPricing() error {
	return m.Err
}

func (m *MockPricingService) IsRegionSupported(region string) bool {
	return true
}

func (m *MockPricingService) GetProducts(ctx context.Context, params *pricing.GetProductsInput, optFns ...func(*pricing.Options)) (*pricing.GetProductsOutput, error) {
	if m.Err != nil {
		return nil, m.Err
	}

	return &pricing.GetProductsOutput{}, nil
}

// CreateTestEC2Instance creates a test EC2 instance for testing
func CreateTestEC2Instance(id string, instanceType ec2types.InstanceType) ec2types.Instance {
	return ec2types.Instance{
		InstanceId:   aws.String(id),
		InstanceType: instanceType,
		State: &ec2types.InstanceState{
			Name: ec2types.InstanceStateNameStopped,
		},
		Tags: []ec2types.Tag{},
	}
}

// CreateTestRDSInstance creates a test RDS instance for testing
func CreateTestRDSInstance(id string, instanceType string) rdstypes.DBInstance {
	now := time.Now()
	return rdstypes.DBInstance{
		DBInstanceIdentifier: aws.String(id),
		DBInstanceClass:      aws.String(instanceType),
		DBInstanceStatus:     aws.String("available"),
		Engine:              aws.String("mysql"),
		EngineVersion:       aws.String("8.0.28"),
		InstanceCreateTime:  &now,
		AllocatedStorage:    aws.Int32(20),
		StorageType:         aws.String("gp2"),
	}
}

// CreateTestVolume creates a test EBS volume for testing
func CreateTestVolume(id string, volumeType ec2types.VolumeType) ec2types.Volume {
	return ec2types.Volume{
		VolumeId:   aws.String(id),
		VolumeType: volumeType,
		State:      ec2types.VolumeStateAvailable,
	}
}

// CreateTestDBSnapshot creates a test RDS snapshot for testing
func CreateTestDBSnapshot(id string, dbInstanceId string) rdstypes.DBSnapshot {
	now := time.Now()
	return rdstypes.DBSnapshot{
		DBSnapshotIdentifier: aws.String(id),
		DBInstanceIdentifier: aws.String(dbInstanceId),
		SnapshotCreateTime:   &now,
		Status:               aws.String("available"),
		AllocatedStorage:     aws.Int32(20),
		Engine:               aws.String("mysql"),
		EngineVersion:        aws.String("8.0.28"),
	}
}
