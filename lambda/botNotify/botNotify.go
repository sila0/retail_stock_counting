package main

import (
	"encoding/json"
	"log"
	"net/url"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/line/line-bot-sdk-go/linebot"
)

// HandleLambdaEvent -
func HandleLambdaEvent(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var msg string
	var err error

	msg = request.PathParameters["msg"]

	if err = Notify(msg); err != nil {
		return ErrorMessage("Got error calling alert, ", err), nil
	}

	b, _ := json.Marshal(map[string]string{"status": "succeed"})

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

// Notify -
func Notify(msg string) error {
	var decodedValue string
	var err error
	var bot = new(linebot.Client)
	var textMessage = new(linebot.TextMessage)

	if bot, err = linebot.New(os.Getenv("CHANNEL_SECRET"), os.Getenv("CHANNEL_TOKEN")); err != nil {
		return err
	}

	if decodedValue, err = url.QueryUnescape(msg); err != nil {
		return err
	}

	textMessage = linebot.NewTextMessage(decodedValue)

	if _, err = bot.PushMessage(os.Getenv("RECEIVER"), textMessage).Do(); err != nil {
		return err
	}

	return nil
}

// ErrorMessage - overwrite response body message
func ErrorMessage(msg string, err error) events.APIGatewayProxyResponse {
	log.Print("enter ErrorMessage function")
	statusCode := 400
	e := []byte{}
	if reqerr, ok := err.(awserr.RequestFailure); ok {
		log.Print(msg, err.(awserr.Error).Message())
		statusCode = reqerr.StatusCode()
		e, _ = json.Marshal(map[string]string{
			"errorMessage:": msg + ", " + err.(awserr.Error).Message(),
			"errorCode:":    err.(awserr.Error).Code(),
			"requestID":     reqerr.RequestID()})
	} else if err != nil {
		log.Print(msg, err)
		e, _ = json.Marshal(map[string]string{"errorMessage:": msg + err.Error()})
	} else {
		log.Print(msg)
		e, _ = json.Marshal(map[string]string{"errorMessage:": msg})
	}
	return events.APIGatewayProxyResponse{
		Headers:    map[string]string{"Content-Type": "application/json", "Access-Control-Allow-Origin": "*"},
		StatusCode: statusCode,
		Body:       string(e),
	}
}
