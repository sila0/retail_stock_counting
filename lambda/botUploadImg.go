package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/line/line-bot-sdk-go/linebot"
)

// RequestInput -
type RequestInput struct {
	ImageBase64 string `json:"imageBase64"`
}

// HandleLambdaEvent -
func HandleLambdaEvent(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var reqInput RequestInput
	var err error

	fmt.Printf("%s", request.Body)
	if err = json.Unmarshal([]byte(request.Body), &reqInput); err != nil {
		return ErrorMessage("Got error calling Unmarshal, ", err), nil
	}

	jpgFileName := ConvBase64toJpg(reqInput.ImageBase64)

	if err = UploadImgToS3(jpgFileName); err != nil {
		return ErrorMessage("Got error calling UploadImgToS3, ", err), nil
	}

	//urlStr := GetPresignURL()
	urlStr := "https://d2srcgu60u9n1z.cloudfront.net/" + jpgFileName

	SendLineImage(urlStr)

	b, _ := json.Marshal(map[string]string{"urlStr": urlStr})

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

// ConvBase64toJpg - convert base64 to jpg
func ConvBase64toJpg(data string) string {
	var err error
	var jpgFilename = strconv.Itoa(rand.Intn(10000)) + ".jpg"

	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(data))
	m, formatString, err := image.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}

	bounds := m.Bounds()
	fmt.Println("base64toJpg", bounds, formatString)

	//Encode from image format to writer

	f, err := os.OpenFile("/tmp/"+jpgFilename, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	err = jpeg.Encode(f, m, &jpeg.Options{Quality: 75})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Jpg file", jpgFilename, "created")
	return jpgFilename
}

// UploadImgToS3 -
func UploadImgToS3(fileName string) error {
	sess := session.Must(session.NewSession())
	uploader := s3manager.NewUploader(sess)

	f, err := os.Open("/tmp/" + fileName)
	if err != nil {
		return fmt.Errorf("failed to open file %q, %v", fileName, err)
	}
	// Upload the file to S3.
	_, err = uploader.Upload(
		&s3manager.UploadInput{
			Bucket: aws.String("linebot.sila"),
			Key:    aws.String(fileName),
			Body:   f,
		})
	if err != nil {
		return fmt.Errorf("failed to upload file, %v", err)
	}

	return nil
}

// GetPresignURL -
func GetPresignURL() string {
	sess := session.Must(session.NewSession())

	svc := s3.New(sess)

	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String("linebot.sila"),
		Key:    aws.String("stock.jpg"),
	})

	fmt.Print("time: ", time.Duration(15)*time.Minute)
	urlStr, err := req.Presign(time.Duration(15) * time.Minute)

	if err != nil {
		log.Println("Failed to sign request", err)
	}

	log.Println("The URL is", urlStr)
	return urlStr
}

// SendLineImage -
func SendLineImage(url string) {
	var err error
	var bot = new(linebot.Client)
	var imageMessage = new(linebot.ImageMessage)

	if bot, err = linebot.New(os.Getenv("CHANNEL_SECRET"), os.Getenv("CHANNEL_TOKEN")); err != nil {
		log.Fatal(err)
	}

	imageMessage = linebot.NewImageMessage(url, url)

	if _, err = bot.PushMessage(os.Getenv("RECEIVER"), imageMessage).Do(); err != nil {
		log.Fatal(err)
	}
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
