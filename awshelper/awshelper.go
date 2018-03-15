package awshelper

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

type AWSHelper struct {
	mySession *session.Session
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
