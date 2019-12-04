package main

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
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

func internalServerError() (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
		},
	}, nil
}

func Handler() (interface{}, error) {

	sess, err := session.NewSession()
	if err != nil {
		return internalServerError()
	}

	svc := dynamodb.New(sess)

	getQuery := &dynamodb.QueryInput{
		TableName: aws.String("threads"),
		IndexName: aws.String("part-updated_at-index"),
		ExpressionAttributeNames: map[string]*string{
			"#part": aws.String("part"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":part": {
				N: aws.String("0"),
			},
		},
		KeyConditionExpression: aws.String("#part = :part"),
		ProjectionExpression:   aws.String("id, title, created_at, updated_at"),
		ScanIndexForward:       aws.Bool(false),
		Limit:                  aws.Int64(100),
	}

	result, err := svc.Query(getQuery)
	if err != nil {
		return internalServerError()
	}

	threads := make([]Thread, 0)
	if err := dynamodbattribute.UnmarshalListOfMaps(result.Items, &threads); err != nil {
		return internalServerError()
	}

	jsonBody, err := json.Marshal(threads)
	if err != nil {
		return internalServerError()
	}

	return events.APIGatewayProxyResponse{
		Body:       string(jsonBody),
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
		},
	}, nil
}

func main() {
	lambda.Start(Handler)
}
