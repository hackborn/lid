package dlock

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"time"
)

// ------------------------------------------------------------
// AWS-SERVICE

// awsService provides a Service implementation on AWS DynamoDB.
// This is basically the point of this package, so it's the
// one and only service.
type awsService struct {
	db *dynamodb.DynamoDB
}

func NewAwsServiceFromSession(sess *session.Session) (Service, error) {
	return _newAwsServiceFromSession(sess)
}

func _newAwsServiceFromSession(sess *session.Session) (*awsService, error) {
	if sess == nil {
		return nil, sessionRequiredErr
	}
	db := dynamodb.New(sess)
	if db == nil {
		return nil, dynamoRequiredErr
	}
	return &awsService{db: db}, nil
}

func (s *awsService) Lock(req LockRequest, opts *LockOpts) (LockResponse, error) {
	return LockResponse{}, nil
}

// createTable() creates my lock table.
func (s *awsService) createTable(name string) error {
	// Define table
	partitionname := "dsig"
	partitiontype := "S"
	att1 := &dynamodb.AttributeDefinition{
		AttributeName: aws.String(partitionname),
		AttributeType: aws.String(partitiontype),
	}
	ttlname := "dttl"
	/*
		ttltype := "N"
		att2 := &dynamodb.AttributeDefinition{
			AttributeName: aws.String(ttlname),
			AttributeType: aws.String(ttltype),
		}
	*/
	key := &dynamodb.KeySchemaElement{
		AttributeName: aws.String(partitionname),
		KeyType:       aws.String("HASH"),
	}
	params := &dynamodb.CreateTableInput{
		TableName:            aws.String(name),
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
		return err
	}

	// Wait for table to be ready
	cond := func() bool {
		return s.tableStatus(name) == awsReady
	}
	err = wait(cond)
	if err != nil {
		return err
	}

	// Enable time to live
	ttlparams := &dynamodb.UpdateTimeToLiveInput{
		TableName: aws.String(name),
		TimeToLiveSpecification: &dynamodb.TimeToLiveSpecification{
			AttributeName: aws.String(ttlname),
			Enabled:       aws.Bool(true),
		},
	}
	_, err = s.db.UpdateTimeToLive(ttlparams)
	return err
}

// deleteTable() deletes the table with the given name. Obviously this is an incredibly
// dangerous function; it's used by testing but should not be used otherwise.
func (s *awsService) deleteTable(name string) {
	if name == "" {
		panic("awsService.deleteTable() with no table name")
	}
	params := &dynamodb.DeleteTableInput{
		TableName: aws.String(name),
	}
	_, err := s.db.DeleteTable(params)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == dynamodb.ErrCodeResourceNotFoundException {
			return
		}
		panic("Error deleting table: " + err.Error())
	}
	cond := func() bool {
		return s.tableStatus(name) == awsMissing
	}
	mustErr(wait(cond))
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

// ----------------------------------------
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
		return conditionFailedErr
	}
	return nil
}

// ----------------------------------------
// CONST and VAR

type awsTableStatus int

const (
	awsMissing  awsTableStatus = iota // The table does not exist
	awsCreating                       // The table is being created
	awsReady                          // The table is ready
)
