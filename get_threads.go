package main

import (
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Thread struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

type Threads struct {
	Threads []Thread
}

type Response struct {
	Threads []Thread `json:"body"`
}

func Handler() (Response, error) {

	sess, err := session.NewSession()
	if err != nil {
		panic(err)
	}

	svc := dynamodb.New(sess)

	getQuery := &dynamodb.QueryInput{
		TableName: aws.String("threads"),
		ExpressionAttributeNames: map[string]*string{
			"#Type": aws.String("type"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":type": {
				S: aws.String("thr"),
			},
		},
		KeyConditionExpression: aws.String("#Type = :type"),
		ProjectionExpression:   aws.String("id, title, created_at, updated_at"),
		ScanIndexForward:       aws.Bool(true),
		Limit:                  aws.Int64(100),
	}

	result, err := svc.Query(getQuery)
	if err != nil {
		fmt.Println("Query error: ", err)
	}

	threads := make([]Thread, 0)
	if err := dynamodbattribute.UnmarshalListOfMaps(result.Items, &threads); err != nil {
		fmt.Println("Unmarshalable error: ", err)
	}

	return Response{
		Threads: threads,
	}, nil
}

func main() {
	lambda.Start(Handler)
}
