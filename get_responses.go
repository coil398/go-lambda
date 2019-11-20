package main

import (
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"reflect"
)

type Request struct {
	ThreadID string `json:"id"`
}

type Response struct {
	Responses interface{} `json:"body"`
}

func Handler(request Request) (interface{}, error) {

	sess, err := session.NewSession()
	if err != nil {
		panic(err)
	}

	svc := dynamodb.New(sess)

	getQuery := &dynamodb.QueryInput{
		TableName: aws.String("responses"),
		ExpressionAttributeNames: map[string]*string{
			"#ThreadID": aws.String("thread_id"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":id": {
				S: aws.String(request.ThreadID),
			},
		},
		KeyConditionExpression: aws.String("#ThreadID = :id"),
		ProjectionExpression:   aws.String("responses"),
	}

	result, err := svc.Query(getQuery)
	if err != nil {
		fmt.Println("Query error: ", err)
	}

	responses := make([]interface{}, 0)
	if err := dynamodbattribute.UnmarshalListOfMaps(result.Items, &responses); err != nil {
		fmt.Println("Unmarshalable error: ", err)
	}

	rv := reflect.ValueOf(responses[0])
	res := rv.MapIndex(reflect.ValueOf("responses")).Interface()

	return Response{
		Responses: res,
	}, nil
}

func main() {
	lambda.Start(Handler)
}
