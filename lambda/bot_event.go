package main

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// HandleLambdaEvent - update dynamodb
func HandleLambdaEvent(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var err error
	var b []byte
	body := request.Body

	if b, err = json.Marshal(body); err != nil {
		fmt.Println(err)
	}

	fmt.Println(string(b))

	return events.APIGatewayProxyResponse{
		Headers:         map[string]string{"Content-Type": "application/json", "Access-Control-Allow-Origin": "*"},
		StatusCode:      200,
		Body:            string(b),
		IsBase64Encoded: false,
	}, nil

}

func main() {
	lambda.Start(HandleLambdaEvent)
}
