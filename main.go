package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/hailocab/dns-lambda/aws"
	"github.com/hailocab/dns-lambda/lambda"

	"github.com/apex/go-apex"
	"github.com/apex/go-apex/cloudwatch"
)

var (
	configFile string
)

func init() {
	flagSet := flag.NewFlagSet("lambda", flag.ContinueOnError)
	flagSet.StringVar(&configFile, "config-file", lambda.DefaultConfigFile, "Location of config file")
	if err := flagSet.Parse(os.Args[1:]); err != nil {
		os.Exit(1)
	}

	log.SetOutput(os.Stderr)
}

func main() {
	log.Printf("Using log file: %s", configFile)

	cloudwatch.HandleFunc(func(evt *cloudwatch.Event, ctx *apex.Context) error {
		if evt.Source != "aws.autoscaling" {
			return errors.New("Not an autoscaling event")
		}

		var details cloudwatch.AutoScalingGroupDetail
		if err := json.Unmarshal(evt.Detail, &details); err != nil {
			log.Printf("Unable to unmarshal detail body: %v", err)
			return err
		}

		config, err := lambda.LoadConfig(configFile)
		if err != nil {
			log.Printf("Config error: %v", err)
			return err
		}

		if config.CreateIPRecords {
			switch lambda.DetermineAutoScalingEventType(evt.DetailType) {
			case lambda.AutoScalingEventLaunch:
				log.Printf("Creating IP based DNS record for %q", details.EC2InstanceID)
			case lambda.AutoScalingEventTerminate:
				log.Printf("Removing IP based DNS record for %q", details.EC2InstanceID)
			}
		}

		log.Printf("Finding instances for %q in %q", details.AutoScalingGroupName, evt.Region)
		resource, err := aws.FindInstances(details.AutoScalingGroupName, evt.Region)
		if err != nil {
			log.Printf("Unable to find instances for %q: %v", details.AutoScalingGroupName, err)
			return err
		}

		var (
			azIPs  = map[string][]string{}
			allIPs []string
			role   string
		)

		roleRE := regexp.MustCompile(fmt.Sprintf("-%s", config.EnvironmentName))
		role = roleRE.ReplaceAllLiteralString(details.AutoScalingGroupName, "")
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
					"AutoScalingGroup": details.AutoScalingGroupName,
					"Role":             role,
					"AvailabilityZone": az,
					"EnvironmentName":  config.EnvironmentName,
				})

				if err != nil {
					return err
				}

				if len(azIPs[az]) == 0 {
					log.Printf("Tring to delete DNS record for: %s", dns)
					aws.DeleteRecord(config.HostedZoneID, dns)
				} else {
					if err := aws.CreateRecord(config.HostedZoneID, dns, azIPs[az]); err != nil {
						return fmt.Errorf("Unable to create AZ record %q: %v (%v)", dns, azIPs[az], err)
					}
				}
			}
		}

		p, ok := config.Patterns["region"]
		if ok {
			dns, err := p.Parse(map[string]string{
				"AutoScalingGroup": details.AutoScalingGroupName,
				"Role":             role,
				"Region":           evt.Region,
				"EnvironmentName":  config.EnvironmentName,
			})
			if err != nil {
				return err
			}

			if len(allIPs) == 0 {
				log.Printf("Tring to delete DNS record for: %s", dns)
				err := aws.DeleteRecord(config.HostedZoneID, dns)
				log.Printf("Delete: %v", err)
			} else {
				if err := aws.CreateRecord(config.HostedZoneID, dns, allIPs); err != nil {
					return fmt.Errorf("Unable to create region record %q: %v ", dns, allIPs)
				}
			}
		}

		return nil
	})
}
