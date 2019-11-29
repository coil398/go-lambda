package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
)

type Request struct {
	Title   string `json:"title"`
	Name    string `json:"name"`
	Content string `json:"content"`
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

func insertData(svc *dynamodb.DynamoDB, data interface{}, target string, wg *sync.WaitGroup) {

	defer wg.Done()

	av, err := dynamodbattribute.MarshalMap(data)
	if err != nil {
		panic(fmt.Sprintf("marshal error, %v\n", err))
	}

	putParams := &dynamodb.PutItemInput{
		TableName: aws.String(target),
		Item:      av,
	}

	_, putErr := svc.PutItem(putParams)
	if putErr != nil {
		panic(fmt.Sprintf("database put error, %v\n", putErr))
	}
}

func Handler(request Request) (events.APIGatewayProxyResponse, error) {
	// name := request.QueryStringParameters["name"]
	// title := request.QueryStringParameters["title"]
	// content := request.QueryStringParameters["content"]
	name := request.Name
	title := request.Title
	content := request.Content

	sess, err := session.NewSession()
	if err != nil {
		panic(err)
	}

	svc := dynamodb.New(sess)
	threadID := uuid.New().String()

	t := thread{
		Type:      "thr",
		ID:        threadID,
		Title:     title,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	r := responses{
		ThreadID: threadID,
		Responses: []response{
			response{
				Name:      name,
				CreatedAt: time.Now().Unix(),
				Content:   content,
			},
		},
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go insertData(svc, t, "threads", &wg)
	wg.Add(1)
	go insertData(svc, r, "responses", &wg)

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
