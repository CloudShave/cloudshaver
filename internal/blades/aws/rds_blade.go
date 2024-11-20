package awsblades

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	awsinterfaces "github.com/cloudshave/cloudshaver/internal/interfaces/aws"
	internaltypes "github.com/cloudshave/cloudshaver/internal/types"
	"github.com/sirupsen/logrus"
)

// Instance type upgrade paths for RDS cost optimization
var rdsInstanceUpgrades = map[string]string{
	"db.t3.micro":  "db.t4g.micro",
	"db.t3.small":  "db.t4g.small",
	"db.t3.medium": "db.t4g.medium",
	"db.r5.large":  "db.r6g.large",
	"db.r5.xlarge": "db.r6g.xlarge",
	"db.m5.large":  "db.m6g.large",
	"db.m5.xlarge": "db.m6g.xlarge",
}

// RDSBlade implements cost optimization analysis for RDS
type RDSBlade struct {
	rdsClient        awsinterfaces.RDSClientAPI
	cloudWatchClient awsinterfaces.CloudWatchClientAPI
	pricingService   awsinterfaces.PricingServiceAPI
	region           string
}

// NewRDSBlade creates a new RDS blade instance
func NewRDSBlade(rdsClient awsinterfaces.RDSClientAPI, cloudWatchClient awsinterfaces.CloudWatchClientAPI, pricingService awsinterfaces.PricingServiceAPI, region string) (*RDSBlade, error) {
	return &RDSBlade{
		rdsClient:        rdsClient,
		cloudWatchClient: cloudWatchClient,
		pricingService:   pricingService,
		region:           region,
	}, nil
}

// GetName returns the name of the blade
func (b *RDSBlade) GetName() string {
	return "RDS Optimization Blade"
}

// GetCategory returns the category of the blade
func (b *RDSBlade) GetCategory() string {
	return string(internaltypes.DatabaseOptimization)
}

// Execute runs the cost optimization analysis
func (b *RDSBlade) Execute() (*internaltypes.BladeResult, error) {
	// Log the region being analyzed
	logrus.Infof("Starting RDS analysis in region: %s", b.region)

	// Initialize the result
	result := &internaltypes.BladeResult{
		CloudProvider:    string(internaltypes.AWS),
		Category:         string(internaltypes.DatabaseOptimization),
		ResourceType:     "RDS",
		PotentialSavings: 0,
		Recommendations:  []string{},
		Details:          make(map[string]string),
		Timestamp:        time.Now(),
	}

	// Get all RDS instances
	instances, err := b.rdsClient.DescribeDBInstances(context.TODO(), &rds.DescribeDBInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe DB instances: %w", err)
	}

	// Get snapshots for backup analysis
	snapshots, err := b.rdsClient.DescribeDBSnapshots(context.TODO(), &rds.DescribeDBSnapshotsInput{})
	if err != nil {
		logrus.WithError(err).Error("Failed to get DB snapshots")
	}

	// Track total potential savings
	var totalPotentialSavings float64

	// Analyze instances for optimization opportunities
	for _, instance := range instances.DBInstances {
		// Skip Aurora instances as they have different optimization strategies
		if instance.Engine != nil && (*instance.Engine == "aurora" || *instance.Engine == "aurora-mysql" || *instance.Engine == "aurora-postgresql") {
			continue
		}

		// Get instance metrics
		metrics, err := b.getInstanceMetrics(instance)
		if err != nil {
			logrus.WithError(err).Errorf("Failed to get metrics for instance %s", *instance.DBInstanceIdentifier)
			continue
		}

		instanceSavings := 0.0
		var instanceRecommendations []string

		// 1. Instance Type Optimization
		if instance.DBInstanceClass != nil {
			if targetType, ok := rdsInstanceUpgrades[*instance.DBInstanceClass]; ok {
				savings, err := b.pricingService.CalculateInstanceSavings(
					*instance.DBInstanceClass,
					targetType,
					b.region,
				)
				if err != nil {
					logrus.WithError(err).Errorf("Failed to calculate savings for instance %s: %v", *instance.DBInstanceIdentifier, err)
				} else if savings > 0 {
					instanceSavings += savings
					instanceRecommendations = append(instanceRecommendations,
						fmt.Sprintf("Consider upgrading from %s to %s for monthly savings of $%.2f",
							*instance.DBInstanceClass, targetType, savings))
				}
			}
		}

		// 2. Resource Utilization Analysis
		if metrics.CPUUtilization < 40 && metrics.ConnectionCount < (metrics.MaxConnections*0.4) {
			instanceRecommendations = append(instanceRecommendations,
				fmt.Sprintf("Consider downsizing due to low utilization (CPU: %.1f%%, Connections: %.1f%%)",
					metrics.CPUUtilization, (metrics.ConnectionCount/metrics.MaxConnections)*100))
		}

		// 3. Storage Optimization
		if metrics.StorageUtilization < 50 && *instance.AllocatedStorage > 100 {
			instanceRecommendations = append(instanceRecommendations,
				fmt.Sprintf("Consider reducing allocated storage (Current: %d GB, Utilization: %.1f%%)",
					*instance.AllocatedStorage, metrics.StorageUtilization))
		}

		// 4. Memory Analysis
		if metrics.SwapUsage > 50*1024*1024 { // More than 50MB swap usage
			instanceRecommendations = append(instanceRecommendations,
				fmt.Sprintf("High swap usage detected (%.2f MB). Consider upgrading instance memory",
					metrics.SwapUsage/(1024*1024)))
		}

		// 5. Performance Analysis
		if metrics.DiskQueueDepth > 1 {
			instanceRecommendations = append(instanceRecommendations,
				fmt.Sprintf("High disk queue depth (%.2f). Consider using Provisioned IOPS storage",
					metrics.DiskQueueDepth))
		}

		if metrics.ReadLatency > 0.02 || metrics.WriteLatency > 0.02 { // More than 20ms latency
			instanceRecommendations = append(instanceRecommendations,
				fmt.Sprintf("High I/O latency detected (Read: %.2fms, Write: %.2fms). Consider optimizing storage",
					metrics.ReadLatency*1000, metrics.WriteLatency*1000))
		}

		// 6. Network Analysis
		networkThreshold := 100 * 1024 * 1024 // 100 MB/s
		if metrics.NetworkReceive > float64(networkThreshold) || metrics.NetworkTransmit > float64(networkThreshold) {
			instanceRecommendations = append(instanceRecommendations,
				fmt.Sprintf("High network utilization (Receive: %.2f MB/s, Transmit: %.2f MB/s). Consider network optimization",
					metrics.NetworkReceive/(1024*1024), metrics.NetworkTransmit/(1024*1024)))
		}

		// 7. Multi-AZ and Read Replica Analysis
		if instance.MultiAZ != nil && *instance.MultiAZ {
			if metrics.ReadIOPS > (metrics.WriteIOPS * 4) {
				instanceRecommendations = append(instanceRecommendations,
					"Consider using read replicas instead of Multi-AZ for read-heavy workload")
			}
		} else {
			// Check if instance should have Multi-AZ based on workload
			if metrics.WriteIOPS > 1000 || metrics.ConnectionCount > (metrics.MaxConnections*0.7) {
				instanceRecommendations = append(instanceRecommendations,
					"Consider enabling Multi-AZ for high-availability due to heavy workload")
			}
		}

		// 8. Backup Analysis
		if metrics.BackupRetention < 7 {
			instanceRecommendations = append(instanceRecommendations,
				fmt.Sprintf("Low backup retention period (%d days). Consider increasing for better disaster recovery",
					metrics.BackupRetention))
		}

		// Count snapshots for this instance
		snapshotCount := 0
		for _, snapshot := range snapshots.DBSnapshots {
			if *snapshot.DBInstanceIdentifier == *instance.DBInstanceIdentifier {
				snapshotCount++
			}
		}
		if snapshotCount > 30 {
			instanceRecommendations = append(instanceRecommendations,
				fmt.Sprintf("High number of snapshots (%d). Consider implementing a snapshot cleanup policy",
					snapshotCount))
		}

		// 9. Engine-specific Analysis
		if instance.Engine != nil {
			switch *instance.Engine {
			case "mysql", "mariadb":
				if metrics.DeadlockCount > 0 {
					instanceRecommendations = append(instanceRecommendations,
						fmt.Sprintf("Detected %d deadlocks. Consider reviewing application logic and indexing",
							int(metrics.DeadlockCount)))
				}
			case "postgres":
				if metrics.BlockedTransactions > 5 {
					instanceRecommendations = append(instanceRecommendations,
						fmt.Sprintf("High number of blocked transactions (%.2f avg). Review transaction management",
							metrics.BlockedTransactions))
				}
			}
		}

		// 10. Burst Balance Analysis
		if metrics.BurstBalance < 20 {
			instanceRecommendations = append(instanceRecommendations,
				fmt.Sprintf("Low burst balance (%.2f%%). Consider upgrading to a larger instance type",
					metrics.BurstBalance))
		}

		// Add instance recommendations if any were generated
		if len(instanceRecommendations) > 0 {
			result.Recommendations = append(result.Recommendations,
				fmt.Sprintf("Instance %s:", *instance.DBInstanceIdentifier))
			for _, rec := range instanceRecommendations {
				result.Recommendations = append(result.Recommendations, "  - "+rec)
			}
		}

		totalPotentialSavings += instanceSavings
	}

	// 11. Reserved Instance Analysis
	reserved, err := b.rdsClient.DescribeReservedDBInstances(context.TODO(), &rds.DescribeReservedDBInstancesInput{})
	if err != nil {
		logrus.WithError(err).Error("Failed to get reserved DB instances")
	} else {
		activeReserved := 0
		for _, ri := range reserved.ReservedDBInstances {
			if ri.State != nil && *ri.State == "active" {
				activeReserved++
			}
		}
		coverage := float64(activeReserved) / float64(len(instances.DBInstances)) * 100
		result.Details["Reserved Instance Coverage"] = fmt.Sprintf("%.1f%%", coverage)

		if coverage < 80 {
			result.Recommendations = append(result.Recommendations,
				fmt.Sprintf("Low Reserved Instance coverage (%.1f%%). Consider increasing coverage for consistent workloads", coverage))
		}
	}

	// Add total potential savings
	result.PotentialSavings = totalPotentialSavings
	result.Details["Total Monthly Savings"] = fmt.Sprintf("$%.2f", totalPotentialSavings)

	return result, nil
}

type instanceMetrics struct {
	CPUUtilization      float64
	ConnectionCount     float64
	MaxConnections      float64
	StorageUtilization  float64
	ReadIOPS            float64
	WriteIOPS           float64
	ReadLatency         float64
	WriteLatency        float64
	FreeableMemory      float64
	SwapUsage           float64
	NetworkReceive      float64
	NetworkTransmit     float64
	ReplicaLag          float64
	BackupRetention     int
	BurstBalance        float64
	QueueDepth          float64
	DiskQueueDepth      float64
	DeadlockCount       float64
	BlockedTransactions float64
}

func (b *RDSBlade) getInstanceMetrics(instance rdstypes.DBInstance) (*instanceMetrics, error) {
	endTime := time.Now()
	startTime := endTime.Add(-7 * 24 * time.Hour) // Last 7 days

	input := &cloudwatch.GetMetricDataInput{
		StartTime: aws.Time(startTime),
		EndTime:   aws.Time(endTime),
		MetricDataQueries: []types.MetricDataQuery{
			{
				Id: aws.String("cpu"),
				MetricStat: &types.MetricStat{
					Metric: &types.Metric{
						Namespace:  aws.String("AWS/RDS"),
						MetricName: aws.String("CPUUtilization"),
						Dimensions: []types.Dimension{
							{
								Name:  aws.String("DBInstanceIdentifier"),
								Value: instance.DBInstanceIdentifier,
							},
						},
					},
					Period: aws.Int32(3600),
					Stat:   aws.String("Average"),
				},
			},
			{
				Id: aws.String("connections"),
				MetricStat: &types.MetricStat{
					Metric: &types.Metric{
						Namespace:  aws.String("AWS/RDS"),
						MetricName: aws.String("DatabaseConnections"),
						Dimensions: []types.Dimension{
							{
								Name:  aws.String("DBInstanceIdentifier"),
								Value: instance.DBInstanceIdentifier,
							},
						},
					},
					Period: aws.Int32(3600),
					Stat:   aws.String("Average"),
				},
			},
			{
				Id: aws.String("storage"),
				MetricStat: &types.MetricStat{
					Metric: &types.Metric{
						Namespace:  aws.String("AWS/RDS"),
						MetricName: aws.String("FreeStorageSpace"),
						Dimensions: []types.Dimension{
							{
								Name:  aws.String("DBInstanceIdentifier"),
								Value: instance.DBInstanceIdentifier,
							},
						},
					},
					Period: aws.Int32(3600),
					Stat:   aws.String("Average"),
				},
			},
			{
				Id: aws.String("read_iops"),
				MetricStat: &types.MetricStat{
					Metric: &types.Metric{
						Namespace:  aws.String("AWS/RDS"),
						MetricName: aws.String("ReadIOPS"),
						Dimensions: []types.Dimension{
							{
								Name:  aws.String("DBInstanceIdentifier"),
								Value: instance.DBInstanceIdentifier,
							},
						},
					},
					Period: aws.Int32(3600),
					Stat:   aws.String("Average"),
				},
			},
			{
				Id: aws.String("write_iops"),
				MetricStat: &types.MetricStat{
					Metric: &types.Metric{
						Namespace:  aws.String("AWS/RDS"),
						MetricName: aws.String("WriteIOPS"),
						Dimensions: []types.Dimension{
							{
								Name:  aws.String("DBInstanceIdentifier"),
								Value: instance.DBInstanceIdentifier,
							},
						},
					},
					Period: aws.Int32(3600),
					Stat:   aws.String("Average"),
				},
			},
			{
				Id: aws.String("read_latency"),
				MetricStat: &types.MetricStat{
					Metric: &types.Metric{
						Namespace:  aws.String("AWS/RDS"),
						MetricName: aws.String("ReadLatency"),
						Dimensions: []types.Dimension{
							{
								Name:  aws.String("DBInstanceIdentifier"),
								Value: instance.DBInstanceIdentifier,
							},
						},
					},
					Period: aws.Int32(3600),
					Stat:   aws.String("Average"),
				},
			},
			{
				Id: aws.String("write_latency"),
				MetricStat: &types.MetricStat{
					Metric: &types.Metric{
						Namespace:  aws.String("AWS/RDS"),
						MetricName: aws.String("WriteLatency"),
						Dimensions: []types.Dimension{
							{
								Name:  aws.String("DBInstanceIdentifier"),
								Value: instance.DBInstanceIdentifier,
							},
						},
					},
					Period: aws.Int32(3600),
					Stat:   aws.String("Average"),
				},
			},
			{
				Id: aws.String("freeable_memory"),
				MetricStat: &types.MetricStat{
					Metric: &types.Metric{
						Namespace:  aws.String("AWS/RDS"),
						MetricName: aws.String("FreeableMemory"),
						Dimensions: []types.Dimension{
							{
								Name:  aws.String("DBInstanceIdentifier"),
								Value: instance.DBInstanceIdentifier,
							},
						},
					},
					Period: aws.Int32(3600),
					Stat:   aws.String("Average"),
				},
			},
			{
				Id: aws.String("swap_usage"),
				MetricStat: &types.MetricStat{
					Metric: &types.Metric{
						Namespace:  aws.String("AWS/RDS"),
						MetricName: aws.String("SwapUsage"),
						Dimensions: []types.Dimension{
							{
								Name:  aws.String("DBInstanceIdentifier"),
								Value: instance.DBInstanceIdentifier,
							},
						},
					},
					Period: aws.Int32(3600),
					Stat:   aws.String("Average"),
				},
			},
			{
				Id: aws.String("network_receive"),
				MetricStat: &types.MetricStat{
					Metric: &types.Metric{
						Namespace:  aws.String("AWS/RDS"),
						MetricName: aws.String("NetworkReceiveThroughput"),
						Dimensions: []types.Dimension{
							{
								Name:  aws.String("DBInstanceIdentifier"),
								Value: instance.DBInstanceIdentifier,
							},
						},
					},
					Period: aws.Int32(3600),
					Stat:   aws.String("Average"),
				},
			},
			{
				Id: aws.String("network_transmit"),
				MetricStat: &types.MetricStat{
					Metric: &types.Metric{
						Namespace:  aws.String("AWS/RDS"),
						MetricName: aws.String("NetworkTransmitThroughput"),
						Dimensions: []types.Dimension{
							{
								Name:  aws.String("DBInstanceIdentifier"),
								Value: instance.DBInstanceIdentifier,
							},
						},
					},
					Period: aws.Int32(3600),
					Stat:   aws.String("Average"),
				},
			},
			{
				Id: aws.String("burst_balance"),
				MetricStat: &types.MetricStat{
					Metric: &types.Metric{
						Namespace:  aws.String("AWS/RDS"),
						MetricName: aws.String("BurstBalance"),
						Dimensions: []types.Dimension{
							{
								Name:  aws.String("DBInstanceIdentifier"),
								Value: instance.DBInstanceIdentifier,
							},
						},
					},
					Period: aws.Int32(3600),
					Stat:   aws.String("Average"),
				},
			},
			{
				Id: aws.String("disk_queue_depth"),
				MetricStat: &types.MetricStat{
					Metric: &types.Metric{
						Namespace:  aws.String("AWS/RDS"),
						MetricName: aws.String("DiskQueueDepth"),
						Dimensions: []types.Dimension{
							{
								Name:  aws.String("DBInstanceIdentifier"),
								Value: instance.DBInstanceIdentifier,
							},
						},
					},
					Period: aws.Int32(3600),
					Stat:   aws.String("Average"),
				},
			},
		},
	}

	// Add replica lag metric if this is a read replica
	if instance.ReadReplicaSourceDBInstanceIdentifier != nil {
		input.MetricDataQueries = append(input.MetricDataQueries, types.MetricDataQuery{
			Id: aws.String("replica_lag"),
			MetricStat: &types.MetricStat{
				Metric: &types.Metric{
					Namespace:  aws.String("AWS/RDS"),
					MetricName: aws.String("ReplicaLag"),
					Dimensions: []types.Dimension{
						{
							Name:  aws.String("DBInstanceIdentifier"),
							Value: instance.DBInstanceIdentifier,
						},
					},
				},
				Period: aws.Int32(3600),
				Stat:   aws.String("Average"),
			},
		})
	}

	// Add engine-specific metrics
	if instance.Engine != nil {
		switch *instance.Engine {
		case "mysql", "mariadb":
			input.MetricDataQueries = append(input.MetricDataQueries,
				types.MetricDataQuery{
					Id: aws.String("deadlocks"),
					MetricStat: &types.MetricStat{
						Metric: &types.Metric{
							Namespace:  aws.String("AWS/RDS"),
							MetricName: aws.String("Deadlocks"),
							Dimensions: []types.Dimension{
								{
									Name:  aws.String("DBInstanceIdentifier"),
									Value: instance.DBInstanceIdentifier,
								},
							},
						},
						Period: aws.Int32(3600),
						Stat:   aws.String("Sum"),
					},
				})
		case "postgres":
			input.MetricDataQueries = append(input.MetricDataQueries,
				types.MetricDataQuery{
					Id: aws.String("blocked_transactions"),
					MetricStat: &types.MetricStat{
						Metric: &types.Metric{
							Namespace:  aws.String("AWS/RDS"),
							MetricName: aws.String("BlockedTransactions"),
							Dimensions: []types.Dimension{
								{
									Name:  aws.String("DBInstanceIdentifier"),
									Value: instance.DBInstanceIdentifier,
								},
							},
						},
						Period: aws.Int32(3600),
						Stat:   aws.String("Average"),
					},
				})
		}
	}

	output, err := b.cloudWatchClient.GetMetricData(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to get metric data: %w", err)
	}

	metrics := &instanceMetrics{}
	if instance.BackupRetentionPeriod != nil {
		metrics.BackupRetention = int(*instance.BackupRetentionPeriod)
	}

	for _, result := range output.MetricDataResults {
		if len(result.Values) == 0 {
			continue
		}

		// Calculate average value
		var sum float64
		for _, v := range result.Values {
			sum += v
		}
		avg := sum / float64(len(result.Values))

		switch *result.Id {
		case "cpu":
			metrics.CPUUtilization = avg
		case "connections":
			metrics.ConnectionCount = avg
		case "storage":
			totalStorage := float64(*instance.AllocatedStorage) * 1024 * 1024 * 1024 // Convert GB to bytes
			metrics.StorageUtilization = ((totalStorage - avg) / totalStorage) * 100
		case "read_iops":
			metrics.ReadIOPS = avg
		case "write_iops":
			metrics.WriteIOPS = avg
		case "read_latency":
			metrics.ReadLatency = avg
		case "write_latency":
			metrics.WriteLatency = avg
		case "freeable_memory":
			metrics.FreeableMemory = avg
		case "swap_usage":
			metrics.SwapUsage = avg
		case "network_receive":
			metrics.NetworkReceive = avg
		case "network_transmit":
			metrics.NetworkTransmit = avg
		case "replica_lag":
			metrics.ReplicaLag = avg
		case "burst_balance":
			metrics.BurstBalance = avg
		case "disk_queue_depth":
			metrics.DiskQueueDepth = avg
		case "deadlocks":
			metrics.DeadlockCount = avg
		case "blocked_transactions":
			metrics.BlockedTransactions = avg
		}
	}

	// Set max connections based on instance class
	// These are approximate values, actual values may vary by engine and configuration
	if instance.DBInstanceClass != nil {
		switch *instance.DBInstanceClass {
		case "db.t3.micro":
			metrics.MaxConnections = 66
		case "db.t3.small":
			metrics.MaxConnections = 150
		case "db.t3.medium":
			metrics.MaxConnections = 312
		default:
			metrics.MaxConnections = 5000
		}
	}

	return metrics, nil
}
