package main

import (
	"encoding/json"
	"errors"
	"flag"
	// "fmt"
	"log"
	"os"
	// "regexp"

	// "github.com/hailocab/dns-lambda/aws"
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

		return nil
	})
}
