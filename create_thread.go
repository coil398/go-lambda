package main

import (
	"context"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
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
	Part      int    `json:"part"`
	ID        string `json:"id"`
	Title     string `json:"title"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
	Name      string `json:"name"`
}

func internalServerError() (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
		},
	}, nil
}

func insertData(ctx context.Context, svc *dynamodb.DynamoDB, data interface{}, target string) error {

	av, err := dynamodbattribute.MarshalMap(data)
	if err != nil {
		return err
	}

	putParams := &dynamodb.PutItemInput{
		TableName: aws.String(target),
		Item:      av,
	}

	select {
	case <-ctx.Done():
		return nil
	default:
		_, err = svc.PutItem(putParams)
		if err != nil {
			return err
		}
	}

	return nil
}

func handler(request Request) (events.APIGatewayProxyResponse, error) {
	name := request.Name
	title := request.Title
	content := request.Content

	sess, err := session.NewSession()
	if err != nil {
		return internalServerError()
	}

	svc := dynamodb.New(sess)
	threadID := uuid.New().String()

	t := thread{
		Part:      0,
		ID:        threadID,
		Title:     title,
		Name:      name,
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

	eg, ctx := errgroup.WithContext(context.TODO())
	ctx, cancel := context.WithCancel(ctx)

	eg.Go(func() error {
		return insertData(ctx, svc, t, "threads")
	})
	eg.Go(func() error {
		return insertData(ctx, svc, r, "responses")
	})

	if err := eg.Wait(); err != nil {
		cancel()
		return internalServerError()
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
		},
	}, nil
}

func main() {
	lambda.Start(handler)
}
