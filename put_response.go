package main

import (
	"context"
	"fmt"
	"time"

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

func Handler(ctx context.Context, req Request) (string, error) {

	createdAt := time.Now().Unix()

	sess, err := session.NewSession()
	if err != nil {
		panic(err)
	}

	svc := dynamodb.New(sess)

	r := []Response{Response{
		Name:      req.Name,
		CreatedAt: createdAt,
		Content:   req.Content,
	}}

	fmt.Println(r)

	response, err := dynamodbattribute.Marshal(r)
	if err != nil {
		panic(err)
	}

	fmt.Println(response)

	updateParams := &dynamodb.UpdateItemInput{
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

	u, err := svc.UpdateItem(updateParams)
	if err != nil {
		panic(err)
	}

	fmt.Println(u)

	return "Success", nil
}

func main() {
	lambda.Start(Handler)
}