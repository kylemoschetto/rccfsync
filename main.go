package main

import (
	// GoLang packages
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	// Google & sheets stuff
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"

	// AWS stuff
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type configuration struct {
	Key    string
	Secret string
}

func checkError(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func main() {
	// grab the keys
	data, err := ioutil.ReadFile("googleapi.secret.json")
	checkError(err)
	conf, err := google.JWTConfigFromJSON(data, sheets.SpreadsheetsScope)
	checkError(err)

	// create new Google sheet connection to the API
	client := conf.Client(context.TODO())
	srv, err := sheets.New(client)
	checkError(err)

	// Open the actual spreadsheet and grab values
	spreadsheetID := "102EkmTDD9m0shJdlQJA4JpCjHpDs63C2sE2Q_bgbfhk" // you need to grab this from the file
	readRange := "Dashboard!A1:C150"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	checkError(err)

	// Create the output file
	file, err := os.Create("index.html")
	checkError(err)
	if err != nil {
		fmt.Println("Cannot create the file.")
	}
	defer file.Close()

	// Write data out
	if len(resp.Values) == 0 {
		fmt.Println("No data found in the sheet.")
	} else {
		fmt.Println("Found data and printing to file")
		fmt.Fprintf(file, "<html><head><link rel='stylesheet' type='text/css' href='rccf.css'></head><body>")
		for _, row := range resp.Values {
			if len(row) == 0 {
				// fmt.Println("No data found in this row.")
			} else if len(row) == 2 {
				fmt.Fprintf(file, "<p>%s, %s</p>\n", row[0], row[1])
			} else if len(row) == 3 {
				fmt.Fprintf(file, "<p class='%s'>%s, %s</p>\n", row[2], row[0], row[1])
			}
		}
	}
	fmt.Fprintf(file, "</body></html>")

	/*
		// Send data to S3
	*/

	// Grab keys
	data2, err := os.Open("aws.secret.json")
	decoder := json.NewDecoder(data2)
	configuration := configuration{}
	decoder.Decode(&configuration)
	defer data2.Close()
	checkError(err)
	fmt.Println("Opened the aws file successfully")

	awsAccessKeyID := configuration.Key        // you need to grab this from the file
	awsSecretAccessKey := configuration.Secret // you need to grab this from the file
	// fmt.Println("Key from file: ", awsAccessKeyID)
	// fmt.Println("Secret from file: ", awsSecretAccessKey)
	// fmt.Println("configuration file string: ", configuration)
	// fmt.Println("decoder file string: ", decoder)
	token := ""
	creds := credentials.NewStaticCredentials(awsAccessKeyID, awsSecretAccessKey, token)
	_, err = creds.Get()
	checkError(err)
	if err != nil {
		// handle error
		fmt.Println("Could not read credentials.")
	}
	cfg := aws.NewConfig().WithRegion("us-west-2").WithCredentials(creds)
	svc := s3.New(session.New(), cfg)

	file2, err := os.Open("index.html")
	checkError(err)
	if err != nil {
		// handle error
		fmt.Println("Could not open index.html file.")
	}
	defer file2.Close()

	fileInfo, _ := file2.Stat()
	size := fileInfo.Size()
	buffer := make([]byte, size) // read file content to buffer

	file2.Read(buffer)
	fileBytes := bytes.NewReader(buffer)
	fileType := http.DetectContentType(buffer)
	path := "/" + file2.Name()
	params := &s3.PutObjectInput{
		Bucket: aws.String("lifestyle-challenge-rccf"),
		Key:    aws.String(path),
		Body:   fileBytes,
		//ACL: aws.String("public-read"),
		//GrantRead: aws.String("GrantRead"),
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(fileType),
		ACL:           aws.String("public-read"),
	}
	resp2, err := svc.PutObject(params)
	checkError(err)
	if err != nil {
		// handle error
		fmt.Println("Could not create resp2.")
	}
	fmt.Printf("Successful file upload to s3 - response %s", awsutil.StringValue(resp2))

}
