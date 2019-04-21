package lidaws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/hackborn/lid"
)

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
	b.err = lid.MergeErr(b.err, err)
	return b
}

func (b awsBuilder) value(key string, value interface{}) awsBuilder {
	dst, err := b.marshalToMap(key, value, b.values)
	b.values = dst
	b.err = lid.MergeErr(b.err, err)
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
