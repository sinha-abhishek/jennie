package awshelper

import (
	"errors"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type AWSHelper struct {
	mySession *session.Session
	sqs       *sqs.SQS
	queueUrl  string
}

var instance *AWSHelper
var once sync.Once

func GetInstance() *AWSHelper {
	once.Do(func() {
		instance = &AWSHelper{}
		var err error
		instance.mySession, err = initSession()
		if err != nil {
			instance.mySession = nil
			instance.sqs = nil
		} else {
			instance.sqs = sqs.New(instance.mySession)
		}
	})
	return instance
}

func initSession() (*session.Session, error) {
	return session.NewSession(&aws.Config{
		Region:      aws.String("ap-south-1"),
		Credentials: credentials.NewSharedCredentials("", "default"),
	})
}

func (awshelper *AWSHelper) GetSession() (*session.Session, error) {
	var err error
	if awshelper.mySession == nil {
		awshelper.mySession, err = initSession()
	}
	return awshelper.mySession, err
}

func (awshelper *AWSHelper) GetSQS() (*sqs.SQS, error) {
	if awshelper.sqs == nil {
		return nil, errors.New("SQS not inited")
	}
	return awshelper.sqs, nil
}
