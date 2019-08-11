package lidaws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/hackborn/lid"
	"time"
)

// ------------------------------------------------------------
// AWS-SERVICE

// awsService provides a Service implementation on AWS DynamoDB.
// This is basically the point of this package, so it's the
// one and only service.
type awsService struct {
	db   *dynamodb.DynamoDB
	opts lid.ServiceOpts
}

// NewAwsServiceFromSession constructs a new service based on the provide AWS session.
// I will internally manage my own connection to a DynamoDB client.
func NewAwsServiceFromSession(opts lid.ServiceOpts, sess *session.Session) (lid.Service, error) {
	return _newAwsServiceFromSession(opts, sess)
}

func _newAwsServiceFromSession(opts lid.ServiceOpts, sess *session.Session) (*awsService, error) {
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
	s := &awsService{db: db, opts: opts}
	// Make sure the table has been constructed
	err := s.createTable()
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *awsService) Lock(req lid.LockRequest, opts *lid.LockOpts) (lid.LockResponse, error) {
	if !req.IsValid() {
		return lid.LockResponse{}, lid.ErrBadRequest
	}
	now := time.Now()
	endTime := now.Add(s.opts.Duration)
	record := awsRecord{req.Signature, req.Signee, req.Level, endTime.UnixNano(), s.getTtl(opts), endTime}

	// Acquire the lock. See Service.Lock() for the rules.
	b := awsBuilder{condition: awsAcquireLockCond}
	b = b.value(":se", req.Signee).value(":lv", req.Level).value(":ex", now.UnixNano())
	if b.err != nil {
		return lid.LockResponse{}, b.err
	}

	ls, err := s.putItem(record, b)
	if err != nil {
		if err == errConditionFailed {
			return lid.LockResponse{}, lid.ErrForbidden
		}
		return lid.LockResponse{}, err
	}
	resp := lid.LockResponse{lid.LockOk, ""}
	if ls.Signee != "" {
		if ls.Signee == req.Signee {
			resp.Status = lid.LockRenewed
		} else {
			resp.Status = lid.LockTransferred
			resp.PreviousSignee = ls.Signee
		}
	}
	return resp, err
}

func (s *awsService) getTtl(opts *lid.LockOpts) int64 {
	ttl := s.opts.TimeToLive
	if opts != nil && opts.TimeToLive != emptyTtl {
		ttl = opts.TimeToLive
	}
	if ttl != emptyTtl {
		return time.Now().Add(ttl).Unix()
	}
	return 0
}

func (s *awsService) Unlock(req lid.UnlockRequest, opts *lid.UnlockOpts) (lid.UnlockResponse, error) {
	if !req.IsValid() {
		return lid.UnlockResponse{}, lid.ErrBadRequest
	}

	// Release the lock. See Service.Unlock() for the rules.
	b := awsBuilder{condition: awsReleaseLockCond}
	b = b.key(awsSignatureKey, req.Signature).value(":se", req.Signee)
	if b.err != nil {
		return lid.UnlockResponse{}, b.err
	}

	ls, err := s.deleteItem(b)
	if err != nil {
		if err == errConditionFailed {
			return lid.UnlockResponse{}, lid.ErrForbidden
		}
		return lid.UnlockResponse{}, err
	}
	resp := lid.UnlockResponse{lid.UnlockOk}
	if ls.Signature == "" {
		resp.Status = lid.UnlockNoLock
	}
	return resp, nil
}

// Check() answers the state of a the requested lock. An error is answered
// if the lock doesn't exist.
// DO NOT USE THIS FUNCTION. It doesn't have much value, but exists as
// I transition a service to this library.
func (s *awsService) Check(signature string) (lid.CheckResponse, error) {
	if signature == "" {
		return lid.CheckResponse{}, lid.ErrBadRequest
	}
	b := awsBuilder{}
	b = b.key(awsSignatureKey, signature)
	if b.err != nil {
		return lid.CheckResponse{}, b.err
	}
	r, err := s.getItem(b)
	if err == nil && r.Signee != "" {
		return lid.CheckResponse{r.Signee, r.Level}, nil
	}
	return lid.CheckResponse{}, lid.ErrNotFound
}

// getItem() is a convenience wrapper for DynamoDB's PutItem().
func (s *awsService) getItem(b awsBuilder) (awsRecord, error) {
	if s.db == nil {
		return awsRecord{}, errInitializationFailed
	}
	params := &dynamodb.GetItemInput{
		TableName: aws.String(s.opts.Table),
	}
	b.get(params)
	r, err := s.db.GetItem(params)
	if err != nil {
		return awsRecord{}, err
	}
	if len(r.Item) > 0 {
		record := awsRecord{}
		err = dynamodbattribute.UnmarshalMap(r.Item, &record)
		if err == nil {
			return record, nil
		}
	}
	return awsRecord{}, lid.ErrNotFound
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
// CONST and VAR

const (
	awsSignatureKey = "lsig"
	awsSigneeKey    = "lsignee"
	awsLevelKey     = "llevel"
	awsExpiresKey   = "lexpires"
)

var (
	awsEmptyDuration = time.Second * 0
	emptyTtl         time.Duration

	awsAcquireLockCond = `attribute_not_exists(` + awsSignatureKey + `) OR ` + awsSigneeKey + ` = :se OR ` + awsLevelKey + ` < :lv OR ` + awsExpiresKey + ` < :ex`
	awsReleaseLockCond = `attribute_not_exists(` + awsSignatureKey + `) OR ` + awsSigneeKey + ` = :se`
)
