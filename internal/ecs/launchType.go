package ecs

const (
	ecsFargateLaunchType = "fargate"
	ecsEC2LaunchType     = "ec2"
)

// LaunchType represents the current container AWS Launchtype
// it can be "fargate" or "ec2".
type LaunchType string

// NewLaunchType returns the container ECS LaunchType.
func NewLaunchType(fargate bool) LaunchType {
	if fargate {
		return LaunchType(ecsFargateLaunchType)
	}
	return LaunchType(ecsEC2LaunchType)
}
