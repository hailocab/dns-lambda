package aws

import (
	"fmt"
	"net"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

// CreateRecord for hosted zone id
func CreateRecord(zone string, record string, dest []string) error {
	return CreateResourceRecordType(zone, "A", record, dest)
}

// CreateTXTRecord creates a txt record
func CreateTXTRecord(zone string, record string, value string) error {
	return CreateResourceRecordType(zone, "TXT", record, []string{value})
}

// CreateResourceRecordType creates a record of a given type
func CreateResourceRecordType(zone string, recordType string, record string, dest []string) error {
	recs := []*route53.ResourceRecord{}
	for _, ip := range dest {
		recs = append(recs, &route53.ResourceRecord{Value: aws.String(ip)})
	}

	svc := route53.New(session.New(), aws.NewConfig())
	_, err := svc.ChangeResourceRecordSets(&route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("UPSERT"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name:            aws.String(record),
						Type:            aws.String(recordType),
						ResourceRecords: recs,
						TTL:             aws.Int64(60),
					},
				},
			},
			Comment: aws.String("ResourceDescription"),
		},
		HostedZoneId: aws.String(zone),
	})

	return err
}

// DeleteRecord deletes a record
func DeleteRecord(zone string, record string) error {
	values, err := FindRecordValues(zone, record)
	if err != nil {
		return err
	}

	return DeleteResourceRecordType(zone, "A", record, values)
}

// DeleteResourceRecordType deletes a record of a given type
func DeleteResourceRecordType(zone string, recordType string, record string, values []string) error {
	recs := []*route53.ResourceRecord{}
	for _, v := range values {
		recs = append(recs, &route53.ResourceRecord{Value: aws.String(v)})
	}

	svc := route53.New(session.New(), aws.NewConfig())
	_, err := svc.ChangeResourceRecordSets(&route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("DELETE"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name:            aws.String(record),
						Type:            aws.String(recordType),
						ResourceRecords: recs,
						TTL:             aws.Int64(60),
					},
				},
			},
			Comment: aws.String("ResourceDescription"),
		},
		HostedZoneId: aws.String(zone),
	})

	return err
}

// FindRecordValues finds a record's value
func FindRecordValues(zone string, record string) ([]string, error) {
	svc := route53.New(session.New(), aws.NewConfig())
	resp, err := svc.ListResourceRecordSets(&route53.ListResourceRecordSetsInput{
		HostedZoneId:    aws.String(zone),
		StartRecordName: aws.String(record),
	})
	if err != nil {
		return nil, err
	}

	values := []string{}
	for _, set := range resp.ResourceRecordSets {
		if !strings.Contains(*set.Name, record) {
			continue
		}

		for _, rec := range set.ResourceRecords {
			values = append(values, *rec.Value)
		}
	}

	return values, nil
}

// IPRecordConfig type:
type IPRecordConfig struct {
	Zone       string
	Record     string
	Value      string
	InstanceID string
	Domain     string
	Region     string
}

// CreateIPRecord creates an IP record
func CreateIPRecord(config *IPRecordConfig) error {
	if err := CreateResourceRecordType(config.Zone, "A", config.Record, []string{config.Value}); err != nil {
		return err
	}

	txt := fmt.Sprintf("%s.%s", config.InstanceID, config.Domain)

	return CreateResourceRecordType(config.Zone, "TXT", txt, []string{config.Record})
}

// DeleteIPRecord creates an IP record
func DeleteIPRecord(config *IPRecordConfig) error {
	txtValue, err := ResolveTXTRecord(fmt.Sprintf("%s.%s.%s", config.InstanceID, config.Region, config.Domain))
	if err != nil {
		return err
	}

	if err := DeleteResourceRecordType(config.Zone, "A", config.Record, txtValue); err != nil {
		return err
	}

	txt := fmt.Sprintf("%s.%s.%s", config.InstanceID, config.Region, config.Domain)

	return DeleteResourceRecordType(config.Zone, "TXT", txt, txtValue)
}

// ResolveTXTRecord resolves a TXT record
func ResolveTXTRecord(record string) ([]string, error) {
	return net.LookupTXT(record)
}
