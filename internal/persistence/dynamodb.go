package persistence

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/connorpalermo/url-shortener/constant/logkey"
	"go.uber.org/zap"
)

type (
	UrlDB struct {
		Logger    *zap.Logger
		DBClient  DBProvider
		TableName string
	}

	DBProvider interface {
		GetItem(context.Context, *dynamodb.GetItemInput, ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
		PutItem(context.Context, *dynamodb.PutItemInput, ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	}
)

const (
	ShortURL      = "short_url"
	ID            = "id"
	OriginalURL   = "original_url"
	DefaultRegion = "us-east-1"
	URLTable      = "url-mapping"
)

func New(logger *zap.Logger) (*UrlDB, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(DefaultRegion),
	)
	if err != nil {
		return nil, err
	}

	db := dynamodb.NewFromConfig(cfg)
	return &UrlDB{
		Logger:    logger,
		DBClient:  db,
		TableName: URLTable,
	}, nil
}

func (db *UrlDB) GetItemByPK(shortUrl string) (*dynamodb.GetItemOutput, error) {
	input := &dynamodb.GetItemInput{
		TableName: &db.TableName,
		Key: map[string]types.AttributeValue{
			ShortURL: &types.AttributeValueMemberS{Value: shortUrl},
		},
	}
	result, err := db.DBClient.GetItem(context.Background(), input)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (db *UrlDB) WriteItem(id int64, shortUrl, originalUrl string) error {
	idStr := aws.String(fmt.Sprintf("%d", id))

	item := map[string]types.AttributeValue{
		ID:          &types.AttributeValueMemberN{Value: *idStr},
		ShortURL:    &types.AttributeValueMemberS{Value: shortUrl},
		OriginalURL: &types.AttributeValueMemberS{Value: originalUrl},
	}
	input := &dynamodb.PutItemInput{
		TableName: &db.TableName,
		Item:      item,
	}

	_, err := db.DBClient.PutItem(context.Background(), input)
	if err != nil {
		return err
	}
	db.Logger.Info("successfully created database entry for the following values:", zap.String(logkey.ID, *idStr),
		zap.String(logkey.OriginalURL, originalUrl), zap.String(logkey.ShortenedURL, shortUrl))
	return nil
}
