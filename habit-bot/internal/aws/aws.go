package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
)

type AwsWrapper struct {
	session *session.Session
}

func NewAwsWrapper() *AwsWrapper {
	session := session.Must(session.NewSession())
	return &AwsWrapper{
		session: session,
	}
}

func (w *AwsWrapper) GetDynamo() *dynamo.DB {
	return dynamo.New(w.session, &aws.Config{Region: aws.String("eu-central-1")})
}
