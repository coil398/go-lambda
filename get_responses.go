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

	"reflect"
)

type Request struct {
	ThreadID string `json:"id"`
}

type Response struct {
	Responses interface{} `json:"body"`
}

func internalServerError() (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
		},
	}, nil
}

func Handler(request events.APIGatewayProxyRequest) (interface{}, error) {

	var id string
	if tmp, ok := request.PathParameters["id"]; ok == true {
		id = tmp
	} else if tmp, ok := request.QueryStringParameters["id"]; ok == true {
		id = tmp
	} else {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Headers: map[string]string{
				"Access-Control-Allow-Origin": "*",
			},
		}, nil
	}

	sess, err := session.NewSession()
	if err != nil {
		return internalServerError()
	}

	svc := dynamodb.New(sess)

	getItemInput := &dynamodb.GetItemInput{
		TableName: aws.String("responses"),
		Key: map[string]*dynamodb.AttributeValue{
			"thread_id": {
				S: aws.String(id),
			},
		},
	}

	result, err := svc.GetItem(getItemInput)
	if err != nil {
		return internalServerError()
	}

	var responses interface{}
	if err := dynamodbattribute.UnmarshalMap(result.Item, &responses); err != nil {
		return internalServerError()
	}

	rv := reflect.ValueOf(responses)
	res := rv.MapIndex(reflect.ValueOf("responses")).Interface()
	jsonBody, err := json.Marshal(res)
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
