package main

import (
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Request struct {
	ThreadID string `json:"id"`
	Name     string `json:"name"`
	Content  string `json:"content"`
}

type Response struct {
	Name      string `json:"name"`
	CreatedAt int64  `json:"created_at"`
	Content   string `json:"content"`
}

type UpdatedAt struct {
	UpdatedAt int64 `json:"updated_at"`
}

func updateData(svc *dynamodb.DynamoDB, data *dynamodb.UpdateItemInput, wg *sync.WaitGroup) {

	defer wg.Done()

	_, err := svc.UpdateItem(data)
	if err != nil {
		panic(err)
	}
}

func Handler(req Request) (events.APIGatewayProxyResponse, error) {

	now := time.Now().Unix()

	sess, err := session.NewSession()
	if err != nil {
		panic(err)
	}

	svc := dynamodb.New(sess)

	r := []Response{Response{
		Name:      req.Name,
		CreatedAt: now,
		Content:   req.Content,
	}}

	response, err := dynamodbattribute.Marshal(r)
	if err != nil {
		panic(err)
	}

	rInputParams := &dynamodb.UpdateItemInput{
		TableName: aws.String("responses"),
		Key: map[string]*dynamodb.AttributeValue{
			"thread_id": {S: aws.String(req.ThreadID)},
		},
		UpdateExpression: aws.String("SET #ri = list_append(#ri, :vals)"),
		ExpressionAttributeNames: map[string]*string{
			"#ri": aws.String("responses"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":vals": response,
		},
	}

	t := UpdatedAt{
		UpdatedAt: now,
	}

	thread, err := dynamodbattribute.Marshal(t)
	if err != nil {
		panic(err)
	}

	tInputParams := &dynamodb.UpdateItemInput{
		TableName: aws.String("threads"),
		Key: map[string]*dynamodb.AttributeValue{
			"type": {S: aws.String("thr")},
			"id":   {S: aws.String(req.ThreadID)},
		},
		UpdateExpression: aws.String("SET #ri = :vals"),
		ExpressionAttributeNames: map[string]*string{
			"#ri": aws.String("updated_at"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":vals": thread,
		},
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go updateData(svc, rInputParams, &wg)
	wg.Add(1)
	go updateData(svc, tInputParams, &wg)
	wg.Wait()

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
		},
	}, nil
}

func main() {
	lambda.Start(Handler)
}
