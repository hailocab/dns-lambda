package aws

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

// CreateRecord for hosted zone id
func CreateRecord(zone string, record string, dest []string) error {
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
						Type:            aws.String("A"),
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

func DeleteRecord(zone string, record string) error {
	values, err := FindRecordValues(zone, record)
	if err != nil {
		return err
	}

	recs := []*route53.ResourceRecord{}
	for _, v := range values {
		recs = append(recs, &route53.ResourceRecord{Value: aws.String(v)})
	}

	svc := route53.New(session.New(), aws.NewConfig())
	_, err = svc.ChangeResourceRecordSets(&route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("DELETE"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name:            aws.String(record),
						Type:            aws.String("A"),
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
