package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// DescribeAvailabilityZones returns an array of availability zones for a region
func DescribeAvailabilityZones(region string) ([]string, error) {
	svc := ec2.New(session.New(), aws.NewConfig().WithRegion(region))
	resp, err := svc.DescribeAvailabilityZones(&ec2.DescribeAvailabilityZonesInput{})
	if err != nil {
		return []string{}, err
	}

	AZs := []string{}
	for _, az := range resp.AvailabilityZones {
		AZs = append(AZs, *az.ZoneName)
	}

	return AZs, nil
}

func FindInstances(asgName string, region string) (*Resource, error) {}

// FindHealthyAutoScalingGroupInstances returns a list of healthy instances for an ASG
func FindHealthyAutoScalingGroupInstances(name string, region string) ([]string, error) {
	svc := autoscaling.New(session.New(), aws.NewConfig().WithRegion(region))
	resp, err := svc.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String(name)},
	})

	if err != nil {
		return nil, err
	}

	instances := []string{}
	for _, asg := range resp.AutoScalingGroups {
		for _, i := range asg.Instances {
			if *i.HealthStatus == "Healthy" {
				instances = append(instances, *i.InstanceId)
			}
		}
	}

	return instances, nil
}

// HydrateInstances and hydrate them to a useful data structure
func HydrateInstances(instances []string, region string) (*Resource, error) {
	azs, err := DescribeAvailabilityZones(region)
	if err != nil {
		return nil, err
	}

	azInstances := map[string][]*Instance{}
	for _, az := range azs {
		azInstances[az] = []*Instance{}
	}

	if len(instances) == 0 {
		return &Resource{
			InstancesByAvailabilityZone: azInstances,
		}, nil
	}

	is := []*string{}
	for _, i := range instances {
		is = append(is, aws.String(i))
	}

	svc := ec2.New(session.New(), aws.NewConfig().WithRegion(region))
	resp, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: is,
	})

	if err != nil {
		return nil, err
	}

	for _, r := range resp.Reservations {
		for _, i := range r.Instances {
			if i.PrivateIpAddress == nil {
				continue
			}

			az := *i.Placement.AvailabilityZone
			azInstances[az] = append(azInstances[az], &Instance{
				InstanceID:       *i.InstanceId,
				PrivateIPAddress: *i.PrivateIpAddress,
			})
		}
	}

	return &Resource{
		InstancesByAvailabilityZone: azInstances,
	}, nil
}

type Resource struct {
	InstancesByAvailabilityZone map[string][]*Instance
}

type Instance struct {
	InstanceID       string
	PrivateIPAddress string
}
