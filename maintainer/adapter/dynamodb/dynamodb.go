package dynamodb

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/fasthall/gochariots/record"
)

const TABLE_NAME string = "gochariots"

var sess = session.Must(session.NewSession(&aws.Config{Region: aws.String("us-west-2")}))
var svc = dynamodb.New(sess)

func PutRecord(r record.Record) error {
	av, err := dynamodbattribute.MarshalMap(r)
	if err != nil {
		return err
	}
	_, err = svc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(TABLE_NAME),
		Item:      av,
	})
	if err != nil {
		return err
	}
	return nil
}

func PutRecords(records []record.Record) error {
	for _, r := range records {
		err := PutRecord(r)
		if err != nil {
			return err
		}
	}
	return nil
}
