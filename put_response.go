package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"golang.org/x/sync/errgroup"
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

func internalServerError() (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
		},
	}, nil
}

func updateData(ctx context.Context, svc *dynamodb.DynamoDB, data *dynamodb.UpdateItemInput) error {

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("Update data in the table, %v failed", aws.StringValue(data.TableName))
		default:
			_, err := svc.UpdateItem(data)
			if err != nil {
				continue
			}
			return nil
		}
	}
}

func handler(req Request) (events.APIGatewayProxyResponse, error) {

	now := time.Now().Unix()

	sess, err := session.NewSession()
	if err != nil {
		return internalServerError()
	}

	svc := dynamodb.New(sess)

	r := []Response{Response{
		Name:      req.Name,
		CreatedAt: now,
		Content:   req.Content,
	}}

	threadID := req.ThreadID

	response, err := dynamodbattribute.Marshal(r)
	if err != nil {
		return internalServerError()
	}

	rInputParams := &dynamodb.UpdateItemInput{
		TableName: aws.String("responses"),
		Key: map[string]*dynamodb.AttributeValue{
			"thread_id": {S: aws.String(threadID)},
		},
		UpdateExpression: aws.String("SET #ri = list_append(#ri, :vals)"),
		ExpressionAttributeNames: map[string]*string{
			"#ri": aws.String("responses"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":vals": response,
		},
	}

	updatedAt, err := dynamodbattribute.Marshal(now)
	if err != nil {
		return internalServerError()
	}

	tInputParams := &dynamodb.UpdateItemInput{
		TableName: aws.String("threads"),
		Key: map[string]*dynamodb.AttributeValue{
			"part": {N: aws.String("0")},
			"id":   {S: aws.String(threadID)},
		},
		UpdateExpression: aws.String("SET #ri = :vals"),
		ExpressionAttributeNames: map[string]*string{
			"#ri": aws.String("updated_at"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":vals": updatedAt,
		},
	}

	eg, ctx := errgroup.WithContext(context.Background())
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	eg.Go(func() error {
		return updateData(ctx, svc, rInputParams)
	})
	eg.Go(func() error {
		return updateData(ctx, svc, tInputParams)
	})

	if err := eg.Wait(); err != nil {
		return internalServerError()
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
		},
	}, nil
}

func main() {
	lambda.Start(handler)
}
