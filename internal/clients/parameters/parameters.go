package parameters

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

type ParamClient struct {
	store *ssm.SSM
}

func NewParamClient(region string) (*ParamClient, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String(region)},
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, err
	}

	ssmsvc := ssm.New(sess, aws.NewConfig().WithRegion(region))
	return &ParamClient{store: ssmsvc}, nil
}

func (p *ParamClient) GetStringParam(param string) (string, error) {
	output, err := p.store.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(param),
		WithDecryption: aws.Bool(false),
	})
	if err != nil {
		return "", err
	}
	return *output.Parameter.Value, nil
}
