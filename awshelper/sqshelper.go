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
				"DelaySeconds":           aws.String("800"),
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

func SendUpdateMessage(title string, message string, delaySeconds int64) error {
	awsH := GetInstance()
	sqsH, err := awsH.GetSQS()
	qURL := awsH.queueUrl
	if err != nil {
		log.Println("failed to get sqs")
		return err
	}
	result, err := sqsH.SendMessage(&sqs.SendMessageInput{
		DelaySeconds: aws.Int64(delaySeconds),
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

func GetUpdateMessages(title string) ([]*sqs.Message, error) {
	awsH := GetInstance()
	sqsH, err := awsH.GetSQS()
	qURL := awsH.queueUrl
	if err != nil {
		log.Println("failed to get sqs")
		return nil, err
	}
	result, err := sqsH.ReceiveMessage(&sqs.ReceiveMessageInput{
		AttributeNames: []*string{
			aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
		},
		MessageAttributeNames: []*string{
			aws.String(sqs.QueueAttributeNameAll),
		},
		QueueUrl:            &qURL,
		MaxNumberOfMessages: aws.Int64(10),
		VisibilityTimeout:   aws.Int64(36000), // 10 hours
		WaitTimeSeconds:     aws.Int64(0),
	})

	if err != nil {
		log.Println("Error", err)
		return nil, err
	}

	if len(result.Messages) == 0 {
		log.Println("Received no messages")
		return nil, nil
	}
	log.Println("result =", result)

	return result.Messages, nil
}

func DeleteMessages(handles []*string) {
	awsH := GetInstance()
	sqsH, err := awsH.GetSQS()
	qURL := awsH.queueUrl
	if err != nil {
		log.Println("failed to get sqs")
	}
	for _, v := range handles {
		resultDelete, err2 := sqsH.DeleteMessage(&sqs.DeleteMessageInput{
			QueueUrl:      &qURL,
			ReceiptHandle: v,
		})

		if err2 != nil {
			log.Println("Delete Error", err2)
			return
		}
		log.Println("Message Deleted", resultDelete)
	}

}
