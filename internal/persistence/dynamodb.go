package persistence

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"go.uber.org/zap"
)

type (
	UrlDB struct {
		log *zap.Logger
	}

	DBProvider interface {
		GetItem(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error)
	}
)
