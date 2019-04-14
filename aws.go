package tack

import (
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

// NewAwsServiceFromSession constructs a new service based on the provide AWS session.
// I will internally manage my own connection to a DynamoDB client.
func NewAwsServiceFromSession(opts ServiceOpts, sess *session.Session) (Service, error) {
	return _newAwsServiceFromSession(opts, sess)
}

func _newAwsServiceFromSession(opts ServiceOpts, sess *session.Session) (*awsService, error) {
	if sess == nil {
		return nil, errSessionRequired
	}
	if opts.Table == "" {
		return nil, errTableRequired
	}
	if opts.Duration == awsEmptyDuration {
		return nil, errDurationRequired
	}
	db := dynamodb.New(sess)
	if db == nil {
		return nil, errDynamoRequired
	}
	return &awsService{db: db, opts: opts}, nil
}

func (s *awsService) Lock(req LockRequest, opts *LockOpts) (LockResponse, error) {
	if !req.isValid() {
		return LockResponse{}, errBadRequest
	}
	now := time.Now()
	endTime := now.Add(s.opts.Duration)
	record := awsRecord{req.Signature, req.Signee, req.Level, endTime.UnixNano(), endTime}

	// Acquire the lock. See Service.Lock() for the rules.
	b := awsBuilder{condition: awsAcquireLockCond}
	b = b.value(":se", req.Signee).value(":lv", req.Level).value(":ex", now.UnixNano())
	if b.err != nil {
		return LockResponse{}, b.err
	}

	ls, err := s.putItem(record, b)
	if err != nil {
		if err == errConditionFailed {
			return LockResponse{}, &Error{Forbidden, forbiddenMsg, nil}
		}
		return LockResponse{}, err
	}
	resp := LockResponse{LockOk, ""}
	if ls.Signee != "" {
		if ls.Signee == req.Signee {
			resp.Status = LockRenewed
		} else {
			resp.Status = LockTransferred
			resp.PreviousSignee = ls.Signee
		}
	}
	return resp, err
}

func (s *awsService) Unlock(req UnlockRequest, opts *UnlockOpts) (UnlockResponse, error) {
	if !req.isValid() {
		return UnlockResponse{}, errBadRequest
	}

	// Release the lock. See Service.Unlock() for the rules.
	b := awsBuilder{condition: awsReleaseLockCond}
	b = b.key(awsSignatureKey, req.Signature).value(":se", req.Signee)
	if b.err != nil {
		return UnlockResponse{}, b.err
	}

	ls, err := s.deleteItem(b)
	if err != nil {
		if err == errConditionFailed {
			return UnlockResponse{}, &Error{Forbidden, forbiddenMsg, nil}
		}
		return UnlockResponse{}, err
	}
	resp := UnlockResponse{UnlockOk}
	if ls.Signature == "" {
		resp.Status = UnlockNoLock
	}
	return resp, nil
}

// putItem is a convenience wrapper for DynamoDB's PutItem().
func (s *awsService) putItem(item interface{}, b awsBuilder) (awsRecord, error) {
	if s.db == nil {
		return awsRecord{}, errInitializationFailed
	}
	atts, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		return awsRecord{}, err
	}
	params := &dynamodb.PutItemInput{
		TableName:    aws.String(s.opts.Table),
		Item:         atts,
		ReturnValues: aws.String("ALL_OLD"),
	}
	b.put(params)
	resp, err := s.db.PutItem(params)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
			return awsRecord{}, errConditionFailed
		}
		return awsRecord{}, err
	}
	if len(resp.Attributes) < 1 {
		return awsRecord{}, nil
	}
	record := awsRecord{}
	err = dynamodbattribute.UnmarshalMap(resp.Attributes, &record)
	if err == nil && record.ExpiresEpoch != 0 {
		record.Expires = time.Unix(0, record.ExpiresEpoch)
	}
	return record, err
}

// deleteItem is a convenience wrapper for DynamoDB's DeleteItem().
func (s *awsService) deleteItem(b awsBuilder) (awsRecord, error) {
	if s.db == nil {
		return awsRecord{}, errInitializationFailed
	}
	params := &dynamodb.DeleteItemInput{
		TableName:    aws.String(s.opts.Table),
		ReturnValues: aws.String("ALL_OLD"),
	}
	b.delete(params)
	resp, err := s.db.DeleteItem(params)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
			return awsRecord{}, errConditionFailed
		}
		return awsRecord{}, err
	}
	if len(resp.Attributes) < 1 {
		return awsRecord{}, nil
	}
	record := awsRecord{}
	err = dynamodbattribute.UnmarshalMap(resp.Attributes, &record)
	return record, err
}

// ------------------------------------------------------------
// TABLE MANAGEMENT

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
	ttlname := "dttl"
	/*
		ttltype := "N"
		att2 := &dynamodb.AttributeDefinition{
			AttributeName: aws.String(ttlname),
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

// ------------------------------------------------------------
// AWS-BUILDER

// awsBuilder is a helper class for building API params.
type awsBuilder struct {
	keys      map[string]*dynamodb.AttributeValue
	condition string
	values    map[string]*dynamodb.AttributeValue
	err       error
}

func (b awsBuilder) key(key string, value interface{}) awsBuilder {
	dst, err := b.marshalToMap(key, value, b.keys)
	b.keys = dst
	b.err = mergeErr(b.err, err)
	return b
}

func (b awsBuilder) value(key string, value interface{}) awsBuilder {
	dst, err := b.marshalToMap(key, value, b.values)
	b.values = dst
	b.err = mergeErr(b.err, err)
	return b
}

func (b awsBuilder) marshalToMap(key string, value interface{}, dst map[string]*dynamodb.AttributeValue) (map[string]*dynamodb.AttributeValue, error) {
	if dst == nil {
		dst = make(map[string]*dynamodb.AttributeValue)
	}
	v, err := dynamodbattribute.Marshal(value)
	if err != nil {
		return nil, err
	}
	dst[key] = v
	return dst, nil
}

func (b awsBuilder) put(dst *dynamodb.PutItemInput) {
	if b.condition != "" {
		dst.ConditionExpression = aws.String(b.condition)
	}
	if len(b.values) > 0 {
		dst.ExpressionAttributeValues = b.values
	}
}

func (b awsBuilder) delete(dst *dynamodb.DeleteItemInput) {
	if len(b.keys) > 0 {
		dst.Key = b.keys
	}
	if b.condition != "" {
		dst.ConditionExpression = aws.String(b.condition)
	}
	if len(b.values) > 0 {
		dst.ExpressionAttributeValues = b.values
	}
}

// ------------------------------------------------------------
// AWS-RECORD

// awsRecord stores a single entry in the lock table.
type awsRecord struct {
	Signature    string    `json:"dsig"`     // The ID for this lock. MUST MATCH awsSignatureKey
	Signee       string    `json:"dsignee"`  // The owner requesting the lock. MUST MATCH awsSigneeKey
	Level        int       `json:"dlevel"`   // The level of lock requested. Leave this at the default 0 if you don't require levels. MUST MATCH awsLevelKey
	ExpiresEpoch int64     `json:"dexpires"` // The time at which this lock expires (epoch). MUST MATCH awsExpiresKey
	Expires      time.Time `json:"-"`        // The time at which this lock expires. Convenience for clients.
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

	awsSignatureKey = "dsig"
	awsSigneeKey    = "dsignee"
	awsLevelKey     = "dlevel"
	awsExpiresKey   = "dexpires"
)

var (
	awsEmptyDuration = time.Second * 0

	awsAcquireLockCond = `attribute_not_exists(` + awsSignatureKey + `) OR ` + awsSigneeKey + ` = :se OR ` + awsLevelKey + ` < :lv OR ` + awsExpiresKey + ` < :ex`
	awsReleaseLockCond = `attribute_not_exists(` + awsSignatureKey + `) OR ` + awsSigneeKey + ` = :se`
)
