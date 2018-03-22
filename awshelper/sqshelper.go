package awshelper

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

const (
	SQS_QUEUE_NAME = "queue_ids"
)

func InitializeQueues() error {
	awsH := GetInstance()
	sqsH, err := awsH.GetSQS()
	if err != nil {
		log.Println("failed to get sqs")
		return err
	}
	result, err2 := sqsH.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(SQS_QUEUE_NAME),
	})
	log.Println("result=", result)
	if err2 != nil || *result.QueueUrl == "" {
		result2, err3 := sqsH.CreateQueue(&sqs.CreateQueueInput{
			QueueName: aws.String(SQS_QUEUE_NAME),
			Attributes: map[string]*string{
				"DelaySeconds":           aws.String("10"),
				"MessageRetentionPeriod": aws.String("86400"),
			},
		})

		log.Printf("result=%v err=%v", result2.QueueUrl, err3)
		if err3 == nil {
			awsH.queueUrl = *result2.QueueUrl
		}
		return err3
	}
	awsH.queueUrl = *result.QueueUrl
	return err2
}

func SendUpdateMessage(title string, message string, delaySeconds int) error {
	awsH := GetInstance()
	sqsH, err := awsH.GetSQS()
	qURL := awsH.queueUrl
	if err != nil {
		log.Println("failed to get sqs")
		return err
	}
	result, err := sqsH.SendMessage(&sqs.SendMessageInput{
		DelaySeconds: aws.Int64(10),
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"Title": &sqs.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(title),
			},
		},
		MessageBody: aws.String(message),
		QueueUrl:    &qURL,
	})

	if err != nil {
		log.Println("Error", err)
		return err
	}

	log.Println("Success", *result.MessageId)
	return nil
}
