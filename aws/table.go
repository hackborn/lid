package lidaws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hackborn/lid"
	"strings"
	"time"
)

// ------------------------------------------------------------
// AWS-SERVICE TABLE MANAGEMENT

// createTable() creates my lock table.
func (s *awsService) createTable() error {
	if s.opts.Table == "" {
		return errTableRequired
	}
	// Define table
	partitiontype := "S"
	att1 := &dynamodb.AttributeDefinition{
		AttributeName: aws.String(awsSignatureKey),
		AttributeType: aws.String(partitiontype),
	}
	/*
		ttltype := "N"
		att2 := &dynamodb.AttributeDefinition{
			AttributeName: aws.String(ttlAttributeName),
			AttributeType: aws.String(ttltype),
		}
	*/
	key := &dynamodb.KeySchemaElement{
		AttributeName: aws.String(awsSignatureKey),
		KeyType:       aws.String("HASH"),
	}
	params := &dynamodb.CreateTableInput{
		TableName:            aws.String(s.opts.Table),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{att1},
		KeySchema:            []*dynamodb.KeySchemaElement{key},
		// Throughput doesn't really matter. Just about everyone should be using
		// autoscaling, which you have to set manually.
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(5),
		},
	}

	// Create table
	_, err := s.db.CreateTable(params)
	if err != nil {
		// Indicates the table already exists.
		if isAwsErrorCode(err, dynamodb.ErrCodeResourceInUseException) {
			return nil
		}
		return err
	}

	// Wait for table to be ready
	cond := func() bool {
		return s.tableStatus(s.opts.Table) == awsReady
	}
	err = wait(cond)
	if err != nil {
		return err
	}

	// Enable time to live
	var emptyDur time.Duration
	if s.opts.TimeToLive != emptyDur {
		ttlparams := &dynamodb.UpdateTimeToLiveInput{
			TableName: aws.String(s.opts.Table),
			TimeToLiveSpecification: &dynamodb.TimeToLiveSpecification{
				AttributeName: aws.String(ttlAttributeName),
				Enabled:       aws.Bool(true),
			},
		}
		_, err = s.db.UpdateTimeToLive(ttlparams)
		// This is returned by dynalite, all we can do is eat it to prevent
		// incompatibilities.
		if strings.HasPrefix(err.Error(), "UnknownOperationException") {
			err = nil
		}
	}
	return err
}

// deleteTable() deletes the table with the given name. Obviously this is an incredibly
// dangerous function; it's used by testing but should not be used otherwise.
func (s *awsService) deleteTable() {
	if s.opts.Table == "" {
		panic("awsService.deleteTable() with no table name")
	}
	params := &dynamodb.DeleteTableInput{
		TableName: aws.String(s.opts.Table),
	}
	_, err := s.db.DeleteTable(params)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == dynamodb.ErrCodeResourceNotFoundException {
			return
		}
		panic("Error deleting table: " + err.Error())
	}
	cond := func() bool {
		return s.tableStatus(s.opts.Table) == awsMissing
	}
	lid.MustErr(wait(cond))
}

// tableStatus() answers the status of the requested table.
func (s *awsService) tableStatus(name string) awsTableStatus {
	params := &dynamodb.DescribeTableInput{
		TableName: aws.String(name),
	}
	r, err := s.db.DescribeTable(params)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == dynamodb.ErrCodeResourceNotFoundException {
			return awsMissing
		}
		panic(err)
	}

	switch *r.Table.TableStatus {
	case "CREATING":
		return awsCreating
	default:
		return awsReady
	}
}

// ------------------------------------------------------------
// BOILERPLATE

func isAwsErrorCode(err error, code string) bool {
	if err == nil {
		return false
	}
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case code:
			return true
		}
	}
	return false
}

// ------------------------------------------------------------
// WAITING

const (
	waitTime = 120 // in seconds
)

type condition func() bool

// wait() waits for the condition to be true, failing
// if waitTime elapses.
func wait(cond condition) error {
	deadline := time.Now().Add(waitTime * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !cond() {
		return errConditionFailed
	}
	return nil
}

// ------------------------------------------------------------
// CONST and VAR

type awsTableStatus int

const (
	awsMissing  awsTableStatus = iota // The table does not exist
	awsCreating                       // The table is being created
	awsReady                          // The table is ready

	ttlAttributeName = "lttl"
)
