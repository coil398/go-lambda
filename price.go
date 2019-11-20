package main

import (
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/costexplorer"
)

// MyPrice structure
type MyPrice struct {
	Price int `json:"Price:"`
}

func reqestPrice() (price int, err error) {
	mySession, err := session.NewSession()
	if err != nil {
		fmt.Println("Session error")
		fmt.Println(err)
		return
	}
	// svc := costexplorer.New(mySession)
	svc := costexplorer.New(mySession, aws.NewConfig().WithRegion("ap-southeast-1"))
	input := &costexplorer.GetCostAndUsageInput{}
	input.SetTimePeriod(&costexplorer.DateInterval{Start: aws.String("2019-11-01"), End: aws.String("2019-11-07")})
	input.SetGranularity("MONTHLY")
	input.SetMetrics(aws.StringSlice([]string{"UnblendedCost"}))
	output, err := svc.GetCostAndUsage(input)
	if err != nil {
		fmt.Println("GetCostAndUsage error")
		fmt.Println(err)
		return
	}
	fmt.Print(output.GoString())
	if err != nil {
		return
	}

	return
}

func getPrice() (MyPrice, error) {
	price, err := reqestPrice()
	if err != nil {
		return MyPrice{Price: -1}, err
	}
	return MyPrice{Price: price}, nil
}

func main() {
	lambda.Start(getPrice)
}
