# CloudShaver

## Overview
CloudShaver is a cost optimization tool designed to analyze and recommend cost-saving strategies across cloud infrastructure.

## Features
- Multi-cloud support (currently AWS, future expansions planned)
- Categorized cost-saving blades
- Detailed cost analysis and recommendations

## Supported Cloud Providers
- [x] AWS
- [ ] Azure (Planned)
- [ ] GCP (Planned)

## Supported AWS Services

### EC2 (Elastic Compute Cloud)
- [x] Instance right-sizing recommendations
- [x] Generation upgrades (e.g., t2 to t3)
- [x] Stopped instance detection
- [x] Under-utilized instance identification
- [x] Real-time pricing data across all regions
- [x] Cost-saving calculations

### EBS (Elastic Block Storage)
- [x] Unattached volume detection
- [x] Volume type optimization (e.g., gp2 to gp3)
- [x] Multi-region pricing support
- [x] Snapshot management recommendations
- [x] IOPS optimization suggestions

### Future Service Support
### RDS (Relational Database Service)
- [ ] Instance right-sizing based on CPU/Memory metrics
- [ ] Aurora serverless conversion opportunities
- [ ] Multi-AZ deployment cost-benefit analysis
- [ ] Reserved instance coverage gaps
- [ ] Storage over-provisioning detection
- [ ] Read replica optimization

### S3 (Simple Storage Service)
- [ ] Intelligent-Tiering adoption opportunities
- [ ] Lifecycle policy recommendations for infrequent access
- [ ] Object version cleanup recommendations
- [ ] Cross-region replication cost analysis
- [ ] S3 analytics activation for optimization insights
- [ ] Large bucket analysis and cost breakdown

### ELB (Elastic Load Balancer)
- [ ] Idle load balancer detection
- [ ] ALB to NLB conversion opportunities
- [ ] Zombie load balancer cleanup
- [ ] Cross-zone load balancing cost analysis
- [ ] SSL certificate expiration monitoring

### Lambda
- [ ] Memory configuration optimization
- [ ] Timeout setting optimization
- [ ] Execution time analysis and recommendations
- [ ] Provisioned concurrency cost-benefit analysis
- [ ] Cold start impact assessment
- [ ] Code package size optimization

### ElastiCache
- [ ] Node type optimization recommendations
- [ ] Reserved node coverage analysis
- [ ] Multi-AZ cost-benefit analysis
- [ ] Unused cache cluster detection
- [ ] Cache hit ratio optimization suggestions

### DynamoDB
- [ ] On-demand vs provisioned capacity analysis
- [ ] Auto-scaling configuration optimization
- [ ] Reserved capacity recommendations
- [ ] Global table cost optimization
- [ ] Unused table detection
- [ ] Backup retention policy optimization

### NAT Gateway
- [ ] Idle NAT gateway detection
- [ ] Traffic pattern analysis
- [ ] Multi-AZ cost optimization
- [ ] VPC endpoint conversion opportunities

### CloudFront
- [ ] Distribution usage patterns analysis
- [ ] Price class optimization
- [ ] Origin request reduction opportunities
- [ ] SSL certificate consolidation
- [ ] Cache hit ratio optimization

### ECS/EKS (Container Services)
- [ ] Container right-sizing recommendations
- [ ] Fargate vs EC2 cost comparison
- [ ] Spot instance opportunities
- [ ] Cluster utilization optimization
- [ ] Reserved instance coverage for container hosts

### Redshift
- [ ] Cluster right-sizing recommendations
- [ ] Reserved node coverage analysis
- [ ] Unused cluster detection
- [ ] AQUA (Advanced Query Accelerator) adoption analysis
- [ ] Concurrency scaling usage optimization

### General Cost Optimization
- [ ] Reserved Instance/Savings Plan coverage gaps
- [ ] Resource tagging compliance
- [ ] Idle resource detection across services
- [ ] Cross-region resource distribution analysis
- [ ] Service limits and quotas monitoring
- [ ] Cost anomaly detection

## Blade Categories
- Compute Optimization
- Storage Optimization
- Network Optimization
- Resource Utilization

## Development Note
This project was fully created with AI assistance, leveraging advanced AI tools for development, testing, and documentation. This demonstrates the potential of AI-driven software development while maintaining high-quality standards and best practices.

## Getting Started
1. Clone the repository
2. Set up AWS credentials
3. Run the application

## Environment Setup
- Go 1.21+
- AWS SDK v2
- Logrus for logging

## Contributing
Please read CONTRIBUTING.md for details on our code of conduct and the process for submitting pull requests.

## License
This project is licensed under the MIT License - see the LICENSE.md file for details.
