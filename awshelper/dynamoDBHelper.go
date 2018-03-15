package awshelper

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func Init() error {
	helper := GetInstance()
	sess, err := helper.GetSession()
	if err != nil {
		log.Println("Can't initialize session")
		return err
	}
	svc := dynamodb.New(sess)
	result, err2 := svc.ListTables(&dynamodb.ListTablesInput{})
	if err2 != nil {
		log.Println("Can't initialize session", err2)
		return err2
	}
	found := false
	for _, table := range result.TableNames {
		if *table == "users" {
			found = true
			break
		}
	}
	if !found {
		err = createTable(svc)
	}
	return err
}

func createTable(svc *dynamodb.DynamoDB) error {
	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("uid"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("uid"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(1),
			WriteCapacityUnits: aws.Int64(1),
		},
		TableName: aws.String("users"),
	}
	_, err := svc.CreateTable(input)

	if err != nil {
		fmt.Println("Got error calling CreateTable:")
		fmt.Println(err.Error())
	}
	return err
}
