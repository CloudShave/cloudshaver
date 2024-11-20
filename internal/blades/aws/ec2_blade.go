package awsblades

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	awsinterfaces "github.com/cloudshave/cloudshaver/internal/interfaces/aws"
	"github.com/cloudshave/cloudshaver/internal/types"
	"github.com/sirupsen/logrus"
)

// Instance type upgrade paths for cost optimization
var instanceUpgrades = map[string]string{
	"t2.micro":  "t3.micro",
	"t2.small":  "t3.small",
	"t2.medium": "t3.medium",
	"m4.large":  "m5.large",
	"m4.xlarge": "m5.xlarge",
	"c4.large":  "c5.large",
	"c4.xlarge": "c5.xlarge",
}

// Volume type upgrade paths for cost optimization
var volumeUpgrades = map[string]string{
	"gp2": "gp3",
	"io1": "io2",
}

type EC2Blade struct {
	ec2Client      awsinterfaces.EC2ClientAPI
	pricingService awsinterfaces.PricingServiceAPI
	region         string
}

func NewEC2Blade(ec2Client awsinterfaces.EC2ClientAPI, pricingService awsinterfaces.PricingServiceAPI, region string) (*EC2Blade, error) {
	return &EC2Blade{
		ec2Client:      ec2Client,
		pricingService: pricingService,
		region:         region,
	}, nil
}

func (b *EC2Blade) GetName() string {
	return "EC2 Optimization Blade"
}

func (b *EC2Blade) GetCategory() string {
	return string(types.ComputeOptimization)
}

func (b *EC2Blade) Execute() (*types.BladeResult, error) {
	// Log the region being analyzed
	logrus.Infof("Starting EC2 analysis in region: %s", b.region)

	// Collect all optimization results
	result := &types.BladeResult{
		CloudProvider:    string(types.AWS),
		Category:         string(types.ComputeOptimization),
		ResourceType:     "EC2",
		PotentialSavings: 0,
		Recommendations:  []string{},
		Details:          make(map[string]string),
		Timestamp:        time.Now(),
	}

	// Get all EBS volumes
	volumes, err := b.ec2Client.DescribeVolumes(context.TODO(), &ec2.DescribeVolumesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe volumes: %w", err)
	}

	// Check for underutilized instances
	underutilizedSavings, underutilizedRecs, err := b.analyzeUnderutilizedInstances()
	if err != nil {
		logrus.WithError(err).Error("Failed to analyze underutilized instances")
	} else {
		result.PotentialSavings += underutilizedSavings
		result.Recommendations = append(result.Recommendations, underutilizedRecs...)
	}

	// Check for stopped instances
	stoppedSavings, stoppedRecs, err := b.analyzeStoppedInstances()
	if err != nil {
		logrus.WithError(err).Error("Failed to analyze stopped instances")
	} else {
		result.PotentialSavings += stoppedSavings
		result.Recommendations = append(result.Recommendations, stoppedRecs...)
	}

	// Check for unattached volumes
	volumeSavings, volumeRecs, err := b.analyzeUnattachedVolumes(context.TODO(), volumes.Volumes)
	if err != nil {
		logrus.WithError(err).Error("Failed to analyze unattached volumes")
	} else {
		result.PotentialSavings += volumeSavings
		result.Recommendations = append(result.Recommendations, volumeRecs...)
	}

	return result, nil
}

func (b *EC2Blade) analyzeUnderutilizedInstances() (float64, []string, error) {
	describeInput := &ec2.DescribeInstancesInput{
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{"running"},
			},
		},
	}

	instancesOutput, err := b.ec2Client.DescribeInstances(context.TODO(), describeInput)
	if err != nil {
		return 0, nil, err
	}

	var totalSavings float64
	var recommendations []string
	instanceSavings := make(map[string]float64)
	instanceRecommendations := make(map[string]string)

	for _, reservation := range instancesOutput.Reservations {
		for _, instance := range reservation.Instances {
			instanceType := string(instance.InstanceType)
			instanceID := *instance.InstanceId

			// Get instance name from tags
			instanceName := instanceID // Default to ID if no name tag
			for _, tag := range instance.Tags {
				if *tag.Key == "Name" {
					instanceName = *tag.Value
					break
				}
			}

			// Log instance details
			logrus.Infof("Found EC2 instance - Name: %s, ID: %s, Type: %s", instanceName, instanceID, instanceType)

			// Check for instance type upgrade opportunities
			if targetType, ok := instanceUpgrades[instanceType]; ok {
				savings, err := b.pricingService.CalculateInstanceSavings(instanceType, targetType, b.region)
				if err != nil {
					logrus.WithError(err).Errorf("Failed to calculate savings for instance %s", instanceID)
					continue
				}

				if savings > 0 {
					instanceSavings[instanceID] = savings
					instanceRecommendations[instanceID] = fmt.Sprintf("Upgrade from %s to %s", instanceType, targetType)
					totalSavings += savings
				}
			}
		}
	}

	// Generate recommendations
	if len(instanceSavings) > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("Found %d instances with optimization opportunities:", len(instanceSavings)))

		for instanceID, savings := range instanceSavings {
			recommendations = append(recommendations,
				fmt.Sprintf("Instance %s: %s (Monthly savings: $%.2f)",
					instanceID, instanceRecommendations[instanceID], savings))
		}
	}

	return totalSavings, recommendations, nil
}

func (b *EC2Blade) analyzeStoppedInstances() (float64, []string, error) {
	describeInput := &ec2.DescribeInstancesInput{
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{"stopped"},
			},
		},
	}

	instancesOutput, err := b.ec2Client.DescribeInstances(context.TODO(), describeInput)
	if err != nil {
		return 0, nil, err
	}

	var stoppedInstances []string
	var potentialSavings float64
	var volumeDetails []string

	for _, reservation := range instancesOutput.Reservations {
		for _, instance := range reservation.Instances {
			instanceID := *instance.InstanceId

			// Get instance name from tags
			instanceName := instanceID // Default to ID if no name tag
			for _, tag := range instance.Tags {
				if *tag.Key == "Name" {
					instanceName = *tag.Value
					break
				}
			}

			// Log stopped instance details
			logrus.Infof("Found stopped EC2 instance - Name: %s, ID: %s, Type: %s",
				instanceName, instanceID, instance.InstanceType)

			// Get all volumes attached to this instance
			volumeInput := &ec2.DescribeVolumesInput{
				Filters: []ec2types.Filter{
					{
						Name:   aws.String("attachment.instance-id"),
						Values: []string{instanceID},
					},
				},
			}

			volumesOutput, err := b.ec2Client.DescribeVolumes(context.TODO(), volumeInput)
			if err != nil {
				log.Printf("Failed to get volumes for instance %s: %v", instanceID, err)
				continue
			}

			var instanceVolumeCost float64
			for _, volume := range volumesOutput.Volumes {
				// Get volume name from tags
				volumeName := *volume.VolumeId // Default to volume ID
				for _, tag := range volume.Tags {
					if *tag.Key == "Name" {
						volumeName = *tag.Value
						break
					}
				}

				// Log attached volume details
				logrus.Infof("Found attached EBS volume - Name: %s, ID: %s, Type: %s, Size: %d GB, Instance: %s",
					volumeName, *volume.VolumeId, volume.VolumeType, *volume.Size, instanceID)

				if !b.pricingService.IsRegionSupported(b.region) {
					log.Printf("Region %s not supported for pricing calculations", b.region)
					continue
				}

				price, err := b.pricingService.GetVolumePrice(string(volume.VolumeType), b.region)
				if err != nil {
					log.Printf("Failed to get price for volume %s: %v", *volume.VolumeId, err)
					continue
				}

				// Calculate monthly cost: price per GB-month * size
				monthlyCost := price * float64(*volume.Size)
				instanceVolumeCost += monthlyCost

				volumeDetails = append(volumeDetails,
					fmt.Sprintf("Instance %s: %s volume of size %d GB costing $%.2f per month",
						instanceID, volume.VolumeType, *volume.Size, monthlyCost))
			}

			stoppedInstances = append(stoppedInstances, instanceID)
			potentialSavings += instanceVolumeCost
		}
	}

	var recommendations []string
	if len(stoppedInstances) > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("Found %d stopped instances that are still incurring EBS costs:", len(stoppedInstances)))
		recommendations = append(recommendations, volumeDetails...)
		recommendations = append(recommendations,
			fmt.Sprintf("Total potential monthly savings: $%.2f", potentialSavings),
			"Consider taking these actions:",
			"- Terminate instances that have been stopped for extended periods",
			"- Create snapshots of important volumes before termination",
			"- Implement automated cleanup of stopped instances after defined period",
			"- Use automated snapshots to recreate volumes when needed")
	}

	return potentialSavings, recommendations, nil
}

func (b *EC2Blade) analyzeUnattachedVolumes(ctx context.Context, volumes []ec2types.Volume) (float64, []string, error) {
	var potentialSavings float64
	var recommendations []string

	// Log the start of volume analysis
	logrus.Infof("Starting unattached EBS volume analysis in region: %s", b.region)

	for _, volume := range volumes {
		if volume.State != ec2types.VolumeStateAvailable {
			continue
		}

		// Get volume name from tags
		volumeName := *volume.VolumeId // Default to volume ID
		for _, tag := range volume.Tags {
			if *tag.Key == "Name" {
				volumeName = *tag.Value
				break
			}
		}

		// Log unattached volume details
		logrus.Infof("Found unattached EBS volume - Name: %s, ID: %s, Type: %s, Size: %d GB, State: %s",
			volumeName, *volume.VolumeId, volume.VolumeType, *volume.Size, volume.State)

		if !b.pricingService.IsRegionSupported(b.region) {
			recommendations = append(recommendations,
				fmt.Sprintf("Unattached volume %s in region %s (pricing not available)",
					aws.ToString(volume.VolumeId), b.region))
			continue
		}

		price, err := b.pricingService.GetVolumePrice(string(volume.VolumeType), b.region)
		if err != nil {
			// Log error but continue with analysis
			log.Printf("Failed to get price for volume %s: %v", aws.ToString(volume.VolumeId), err)
			continue
		}

		monthlyCost := price * float64(*volume.Size) * 24 * 30 // Monthly cost
		potentialSavings += monthlyCost

		recommendations = append(recommendations,
			fmt.Sprintf("Unattached %s volume %s of size %d GB costing approximately $%.2f per month",
				volume.VolumeType, aws.ToString(volume.VolumeId), *volume.Size, monthlyCost))
	}

	return potentialSavings, recommendations, nil
}
