package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
)

type Response struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type Request struct {
	Title    string   `json:"title"`
	Response Response `json:"response"`
}

type response struct {
	Name      string `json:"name"`
	CreatedAt int64  `json:"created_at"`
	Content   string `json:"content"`
}

type responses struct {
	ThreadID  string     `json:"thread_id"`
	Responses []response `json:"responses"`
}

type thread struct {
	Type      string `json:"type"`
	ID        string `json:"id"`
	Title     string `json:"title"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

func inputData(svc *dynamodb.DynamoDB, data interface{}, target string, wg sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	av, err := dynamodbattribute.MarshalMap(data)
	if err != nil {
		panic(err)
	}

	putParams := &dynamodb.PutItemInput{
		TableName: aws.String(target),
		Item:      av,
	}

	_, putErr := svc.PutItem(putParams)
	if putErr != nil {
		panic(fmt.Sprintf("failed, %v", putErr))
	}
}

func Handler(ctx context.Context, req Request) (res string, err error) {
	sess, err := session.NewSession()
	if err != nil {
		panic(err)
	}

	svc := dynamodb.New(sess)
	threadID := uuid.New().String()

	t := thread{
		Type:      "thr",
		ID:        threadID,
		Title:     req.Title,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	r := responses{
		ThreadID: threadID,
		Responses: []response{
			response{
				Name:      req.Response.Name,
				CreatedAt: time.Now().Unix(),
				Content:   req.Response.Content,
			},
		},
	}

	var wg sync.WaitGroup

	go inputData(svc, t, "threads", wg)
	go inputData(svc, r, "responses", wg)

	wg.Wait()

	return "Success", nil
}

func main() {
	lambda.Start(Handler)
}
