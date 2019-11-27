package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
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

func Handler(request events.APIGatewayProxyRequest) (interface{}, error) {

	id := request.QueryStringParameters["id"]

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
				S: aws.String(id),
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
	jsonBody, err := json.Marshal(res)
	if err != nil {
		panic(err)
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
