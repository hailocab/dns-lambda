package main

import (
	"fmt"
	"log"

	"github.com/hailocab/dns-lambda/aws"
	"github.com/hailocab/dns-lambda/cloudwatch"
	"github.com/hailocab/dns-lambda/lambda"

	"github.com/apex/go-apex"
	// "github.com/k0kubun/pp"
)

func main() {
	cloudwatch.HandleFunc(func(evt *cloudwatch.Event, ctx *apex.Context) error {
		config, err := lambda.LoadConfig("examples/config.json")
		if err != nil {
			return err
		}

		// aws.DeleteRecord(config.HostedZone, "h2o-nonflow1-stg.eu-west-1b.i.stg.foobar.com.")
		// return nil

		asgName, ok := evt.Detail.Get("AutoScalingGroupName")
		if !ok {
			return fmt.Errorf("Error finding ASG name")
		}

		log.Printf("Trying to update records for %q", asgName.(string))

		instances, err := aws.FindHealthyAutoScalingGroupInstances(asgName.(string), evt.Region)
		if err != nil {
			return fmt.Errorf("Unable to find healthy instances for %q: %v", asgName.(string), err)
		}

		log.Printf("Instances found for %q: %v", asgName.(string), instances)

		resource, err := aws.FindInstances(instances, evt.Region)
		if err != nil {
			return fmt.Errorf("Unable to find instances %q: %v", asgName.(string), err)
		}

		azIPs := map[string][]string{}
		allIPs := []string{}
		for az, instances := range resource.InstancesByAvailabilityZone {
			azIPs[az] = []string{}

			for _, i := range instances {
				log.Printf("Found instance %q (%s) in %q", i.InstanceID, i.PrivateIPAddress, az)
				azIPs[az] = append(azIPs[az], i.PrivateIPAddress)
				allIPs = append(allIPs, i.PrivateIPAddress)
			}

			p, ok := config.Patterns["az"]
			if ok {
				dns, err := p.Parse(map[string]string{
					"AutoScalingGroup": asgName.(string),
					"AvailabilityZone": az,
					"EnvironmentName":  config.EnvironmentName,
				})

				if err != nil {
					return err
				}

				if len(azIPs[az]) == 0 {
					log.Printf("Tring to delete DNS record for: %s", dns)
					aws.DeleteRecord(config.HostedZone, dns)
				} else {
					if err := aws.CreateRecord(config.HostedZone, dns, azIPs[az]); err != nil {
						return fmt.Errorf("Unable to create AZ record %q: %v (%v)", dns, azIPs[az], err)
					}
				}
			}
		}

		p, ok := config.Patterns["region"]
		if ok {
			dns, err := p.Parse(map[string]string{
				"AutoScalingGroup": asgName.(string),
				"Region":           evt.Region,
				"EnvironmentName":  config.EnvironmentName,
			})
			if err != nil {
				return err
			}

			if len(allIPs) == 0 {
				log.Printf("Tring to delete DNS record for: %s", dns)
				err := aws.DeleteRecord(config.HostedZone, dns)
				log.Printf("Delete: %v", err)
			} else {
				if err := aws.CreateRecord(config.HostedZone, dns, allIPs); err != nil {
					return fmt.Errorf("Unable to create region record %q: %v ", dns, allIPs)
				}
			}
		}

		return nil
	})
}
