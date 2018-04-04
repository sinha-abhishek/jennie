package authhelper

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/sinha-abhishek/jennie/awshelper"
)

type StorageInterface interface {
	Setup() error
	StoreRefreshToken(identifier string, refreshToken string) error
	GetRefreshToken(identifier string) (string, error)
	DeleteRefreshToken(identifier string) error
}

type DynamoStorage struct {
	db *dynamodb.DynamoDB
}

func GetDynamoStorage() (*DynamoStorage, error) {
	inst := awshelper.GetInstance()
	session, err := inst.GetSession()
	if err != nil {
		return nil, err
	}
	db1 := dynamodb.New(session)
	ds := &DynamoStorage{
		db: db1,
	}
	return ds, err
}

func (ds *DynamoStorage) Setup() error {
	res, err := ds.db.ListTables(&dynamodb.ListTablesInput{})
	if err != nil {
		return err
	}
	found := false
	for _, table := range res.TableNames {
		log.Println("tablename found ", *table)
		if strings.Compare(*table, "refresh_table") == 0 {
			found = true
			break
		}
	}
	if found {
		return nil
	}
	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("identifier"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("identifier"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(1),
			WriteCapacityUnits: aws.Int64(1),
		},
		TableName: aws.String("refresh_table"),
	}
	_, err = ds.db.CreateTable(input)

	if err != nil {
		fmt.Println("Got error calling CreateTable:")
		fmt.Println(err.Error())
	}
	return err
}

type TokeStore struct {
	Identifier   string `json:"identifier"`
	RefreshToken string `json:"refresh_token"`
}

func (ds *DynamoStorage) StoreRefreshToken(identifier string, refreshToken string) error {
	item := TokeStore{
		Identifier:   identifier,
		RefreshToken: refreshToken,
	}
	tn := aws.String("refresh_table")
	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		return err
	}
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: tn,
	}
	_, err = ds.db.PutItem(input)
	return err
}

func (ds *DynamoStorage) GetRefreshToken(identifier string) (string, error) {
	result, err := ds.db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("refresh_table"),
		Key: map[string]*dynamodb.AttributeValue{
			"identifier": {
				S: aws.String(identifier),
			},
		},
	})
	if err != nil {
		log.Println("Can't get input", err)
		return "", err
	}
	item := &TokeStore{}
	err = dynamodbattribute.UnmarshalMap(result.Item, item)
	if err != nil {
		log.Println("Can't unmarshal input ", err)
		return "", err
	}
	if item.Identifier != identifier {
		log.Println("Can't find uid ")
		return "", errors.New("user not found")
	}
	return item.RefreshToken, err
}

func (ds *DynamoStorage) DeleteRefreshToken(identifier string) error {
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"identifier": {
				S: aws.String(identifier),
			},
		},
		TableName: aws.String("refresh_table"),
	}

	_, err := ds.db.DeleteItem(input)

	if err != nil {
		log.Println("Got error calling DeleteItem")
		log.Println(err.Error())
		return err
	}
	return nil
}
