package awshelper

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
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
	found := map[string]bool{"users": false, "responded_ids": false}
	for _, table := range result.TableNames {
		log.Println("tablename found ", *table)
		found[*table] = true
	}
	log.Println("found tables=", found)
	for k, v := range found {
		if v == false {
			err = createTable(svc, k)
			if err != nil {
				log.Println("Failed to create table ", k)
				return err
			}
		}
	}

	return err
}

func createTable(svc *dynamodb.DynamoDB, tableName string) error {
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
		TableName: aws.String(tableName),
	}
	_, err := svc.CreateTable(input)

	if err != nil {
		fmt.Println("Got error calling CreateTable:")
		fmt.Println(err.Error())
	}
	return err
}

type UserItem struct {
	Uid      string `json:"uid"`
	UserData string `json:"userdata"`
}

type UserRespondedIds struct {
	Uid          string          `json:"uid"`
	RespondedIds map[string]bool `json:"responded"`
}

func getService() (*dynamodb.DynamoDB, error) {
	helper := GetInstance()
	sess, err := helper.GetSession()
	if err != nil {
		log.Println("Can't initialize session")
		return nil, err
	}
	svc := dynamodb.New(sess)
	return svc, err
}

func SaveUser(uid string, userData string) error {
	svc, err := getService()
	if err != nil {
		log.Println("Can't initialize session")
		return err
	}
	ud := base64.StdEncoding.EncodeToString([]byte(userData))
	item := UserItem{
		Uid:      uid,
		UserData: ud,
	}
	log.Println("item=", item)
	av, err2 := dynamodbattribute.MarshalMap(item)
	if err2 != nil {
		log.Println("Can't marshal item ", err2)
		return err2
	}
	log.Println("item=", av)
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String("users"),
	}
	_, err = svc.PutItem(input)
	if err != nil {
		log.Println("Can't put item", err)
		return err
	}
	return err
}

func FetchUser(uid string) ([]byte, error) {
	svc, err := getService()
	if err != nil {
		log.Println("Can't initialize session")
		return nil, err
	}
	result, err2 := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("users"),
		Key: map[string]*dynamodb.AttributeValue{
			"uid": {
				S: aws.String(uid),
			},
		},
	})
	if err2 != nil {
		log.Println("Can't get input", err2)
		return nil, err2
	}
	item := &UserItem{}
	err = dynamodbattribute.UnmarshalMap(result.Item, item)
	if err != nil {
		log.Println("Can't unmarshal input ", err)
		return nil, err
	}
	if item.Uid != uid {
		log.Println("Can't find uid ")
		return nil, errors.New("user not found")
	}
	ud := item.UserData
	userData, err3 := base64.StdEncoding.DecodeString(ud)
	if err != nil {
		log.Println("Can't decode user ", err3)
	}
	return userData, err3
}

func GetRespondedIdsForUser(uid string) (*UserRespondedIds, error) {
	svc, err := getService()
	if err != nil {
		log.Println("Can't initialize session")
		return nil, err
	}
	result, err2 := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("responded_ids"),
		Key: map[string]*dynamodb.AttributeValue{
			"uid": {
				S: aws.String(uid),
			},
		},
	})
	if err2 != nil {
		log.Println("Can't get user", err2)
		return nil, err2
	}
	item := &UserRespondedIds{}
	err = dynamodbattribute.UnmarshalMap(result.Item, item)
	if err != nil {
		log.Println("Can't unmarshal input ", err)
		return nil, err
	}
	return item, err
}

func storeRespondedItem(item *UserRespondedIds) error {
	svc, err := getService()
	if err != nil {
		log.Println("Can't initialize session")
		return err
	}
	log.Println("item=", item)
	av, err2 := dynamodbattribute.MarshalMap(item)
	if err2 != nil {
		log.Println("Can't marshal item ", err2)
		return err2
	}
	log.Println("item=", av)
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String("responded_ids"),
	}
	_, err = svc.PutItem(input)
	if err != nil {
		log.Println("Can't put item", err)
		return err
	}
	return err
}

func AppendRespondedID(id string, uid string) error {
	item, err := GetRespondedIdsForUser(uid)
	if err != nil {
		log.Println("can't query", err)
		return err
	}
	if item.Uid != uid {
		//create item with Uid and respoodedId
		item.Uid = uid
		item.RespondedIds = map[string]bool{id: true}
	} else {
		item.RespondedIds[id] = true
	}
	return storeRespondedItem(item)
}

func ClearRespondedIds(uid string) error {
	svc, err := getService()
	if err != nil {
		log.Println("Can't initialize session")
		return err
	}
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"uid": {
				S: aws.String(uid),
			},
		},
		TableName: aws.String("responded_ids"),
	}

	_, err = svc.DeleteItem(input)

	if err != nil {
		log.Println("Got error calling DeleteItem")
		log.Println(err.Error())
		return err
	}
	return nil
}
