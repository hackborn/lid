package dlock

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"time"
)

// ------------------------------------------------------------
// AWS-SERVICE

// awsService provides a Service implementation on AWS DynamoDB.
// This is basically the point of this package, so it's the
// one and only service.
type awsService struct {
	db   *dynamodb.DynamoDB
	opts ServiceOpts
}

func NewAwsServiceFromSession(opts ServiceOpts, sess *session.Session) (Service, error) {
	return _newAwsServiceFromSession(opts, sess)
}

func _newAwsServiceFromSession(opts ServiceOpts, sess *session.Session) (*awsService, error) {
	if sess == nil {
		return nil, sessionRequiredErr
	}
	if opts.Table == "" {
		return nil, tableRequiredErr
	}
	if opts.Duration == emptyDuration {
		return nil, durationRequiredErr
	}
	db := dynamodb.New(sess)
	if db == nil {
		return nil, dynamoRequiredErr
	}
	return &awsService{db: db, opts: opts}, nil
}

func (s *awsService) Lock(req LockRequest, opts *LockOpts) (LockResponse, error) {
	if s.db == nil {
		return LockResponse{}, initializationFailedErr
	}
	endTime := time.Now().Add(s.opts.Duration)
	state := LockState{req.Signature, req.Signee, req.Level, endTime, endTime.Unix()}
	// Acquire a lock if:
	// * It does not exist
	// * Or it does, and I own it
	// * Or it does, but my lock is higher
	err := s.putItem(state, "")
	return LockResponse{}, err
}

// createTable() creates my lock table.
func (s *awsService) createTable() error {
	if s.opts.Table == "" {
		return tableRequiredErr
	}
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
	if !s.opts.TimeToLive.IsZero() {
		ttlparams := &dynamodb.UpdateTimeToLiveInput{
			TableName: aws.String(s.opts.Table),
			TimeToLiveSpecification: &dynamodb.TimeToLiveSpecification{
				AttributeName: aws.String(ttlname),
				Enabled:       aws.Bool(true),
			},
		}
		_, err = s.db.UpdateTimeToLive(ttlparams)
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

func (s *awsService) putItem(item interface{}, condition string) error {
	atts, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		return err
	}
	params := &dynamodb.PutItemInput{
		TableName:    aws.String(s.opts.Table),
		Item:         atts,
		ReturnValues: aws.String("ALL_OLD"),
	}
	if condition != "" {
		params.ConditionExpression = aws.String(condition)
	}
	//	if opts != nil && opts.values != nil {
	//		params.ExpressionAttributeValues = opts.values
	//	}
	resp, err := s.db.PutItem(params)
	fmt.Println("put resp", resp)
	return err
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

var (
	emptyDuration = time.Second * 0
)
