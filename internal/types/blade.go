package types

import "time"

// BladeResult represents the output of a cost-saving blade
type BladeResult struct {
	BladeName        string            `json:"blade_name"`
	CloudProvider    string            `json:"cloud_provider"`
	Category         string            `json:"category"`
	ResourceType     string            `json:"resource_type"`
	ResourceID       string            `json:"resource_id"`
	PotentialSavings float64           `json:"potential_savings"`
	Recommendations  []string          `json:"recommendations"`
	Details          map[string]string `json:"details"`

	Timestamp   time.Time `json:"timestamp"`
	MonthlyCost float64   `json:"monthly_cost,omitempty"`
}

// Blade interface defines the contract for cost-saving blades
type Blade interface {
	// Execute runs the cost-saving analysis
	Execute() (*BladeResult, error)

	// GetName returns the name of the blade
	GetName() string

	// GetCategory returns the category of the blade
	GetCategory() string
}

// CloudProvider enum-like structure
type CloudProvider string

const (
	AWS   CloudProvider = "aws"
	Azure CloudProvider = "azure"
	GCP   CloudProvider = "gcp"
)

// BladeCategory defines standard blade categories
type BladeCategory string

const (
	ComputeOptimization   BladeCategory = "compute"
	StorageOptimization   BladeCategory = "storage"
	NetworkOptimization   BladeCategory = "network"
	DatabaseOptimization  BladeCategory = "database"
	ContainerOptimization BladeCategory = "container"
	BladeUnattachedVolume BladeCategory = "unattached_volume"
)

// VolumeState represents the state of an EBS volume
type VolumeState string

const (
	VolumeStateAvailable VolumeState = "available"
	VolumeStateInUse     VolumeState = "in-use"
	VolumeStateDeleted   VolumeState = "deleted"
)
