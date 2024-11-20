package types

// CloudProvider represents a cloud service provider
type CloudProvider string

const (
	AWS   CloudProvider = "aws"
	Azure CloudProvider = "azure"
	GCP   CloudProvider = "gcp"
)

// Blade represents a cost optimization analysis tool
type Blade interface {
	// GetName returns the name of the blade
	GetName() string

	// GetCategory returns the category of resources this blade analyzes
	GetCategory() string

	// Execute runs the blade's analysis and returns the results
	Execute() (*BladeResult, error)
}

// BladeResult contains the results of a blade's analysis
type BladeResult struct {
	// Name of the blade that produced these results
	BladeName string `json:"blade_name"`

	// Category of resources analyzed
	Category string `json:"category"`

	// Total potential cost savings identified
	PotentialSavings float64 `json:"potential_savings"`

	// Detailed findings and recommendations
	Details []string `json:"details"`
}
