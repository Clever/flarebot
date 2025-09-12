package main

import (
	"context"
	"testing"
	// "github.com/aws/aws-sdk-go-v2/aws"
	// "github.com/aws/aws-sdk-go-v2/service/s3"
	// "github.com/golang/mock/gomock"
	// "github.com/stretchr/testify/assert"
)

//go:generate mockgen -package main -destination mock_s3api.go -source main.go S3ClientInterface

type handleInput struct {
	ctx context.Context
}

type handleOutput struct {
	err error
}

type handleTest struct {
	description string
	input       handleInput
	output      handleOutput
}

func TestHandle(t *testing.T) {
	// tests := []handleTest{
	// 	{
	// 		input: handleInput{
	// 			ctx: context.Background(),
	// 		},
	// 		output: handleOutput{
	// 			err: nil,
	// 		},
	// 		mockExpectations: func(s3Client *MockS3ClientInterface) {
	// 			s3Client.EXPECT().GetObject(gomock.Any(), &s3.GetObjectInput{
	// 				Key:    aws.String("foo"),
	// 				Bucket: aws.String("bar"),
	// 			})
	// 		},
	// 	},
	// }
	// for _, test := range tests {
	// 	t.Run(test.description, func(t *testing.T) {
	// 		mockController := gomock.NewController(t)
	// 		defer mockController.Finish()
	// 		mockS3Client := NewMockS3ClientInterface(mockController)
	// 		test.mockExpectations(mockS3Client)
	// 		err := Handler{s3Client: mockS3Client}.Handle(test.input.ctx, test.input.input)
	// 		assert.Equal(t, test.output.err, err)
	// 	})
	// }
}
