package persistence

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockDynamoDBClient struct {
	mock.Mock
}

func (m *MockDynamoDBClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*dynamodb.GetItemOutput), args.Error(1)
}

func (m *MockDynamoDBClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*dynamodb.PutItemOutput), args.Error(1)
}

var tableName = "url-mapping"

func Test_CreateClient(t *testing.T) {
	logger, _ := zap.NewProduction()

	New(logger)
}

func Test_GetItemByPK(t *testing.T) {
	tests := map[string]struct {
		shortUrl     string
		input        *dynamodb.GetItemInput
		output       *dynamodb.GetItemOutput
		getItemError error
		checkError   bool
	}{
		"GetItemByPK Happy Path": {
			shortUrl: "b",
			input: &dynamodb.GetItemInput{
				TableName: &tableName,
				Key: map[string]types.AttributeValue{
					ShortURL: &types.AttributeValueMemberS{Value: "b"},
				},
			},
			output: &dynamodb.GetItemOutput{
				Item: map[string]types.AttributeValue{
					"id":           &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", int64(1))},
					"short_url":    &types.AttributeValueMemberS{Value: "b"},
					"original_url": &types.AttributeValueMemberS{Value: "http://www.example.com"},
				},
			},
		},
		"GetItemByPK Sad Path": {
			shortUrl: "12412352",
			input: &dynamodb.GetItemInput{
				TableName: &tableName,
				Key: map[string]types.AttributeValue{
					ShortURL: &types.AttributeValueMemberS{Value: "12412352"},
				},
			},
			checkError:   true,
			getItemError: errors.New("error"),
		},
	}
	logger, _ := zap.NewProduction()

	for _, tc := range tests {
		m := &MockDynamoDBClient{}

		m.On("GetItem", context.Background(), tc.input).Return(tc.output, tc.getItemError)

		db := &UrlDB{
			Logger:    logger,
			DBClient:  m,
			TableName: URLTable,
		}

		output, err := db.GetItemByPK(tc.shortUrl)

		if tc.checkError {
			assert.Error(t, err)
			return
		}

		assert.Equal(t, output, tc.output)
	}
}

func Test_WriteItem(t *testing.T) {
	tests := map[string]struct {
		shortUrl       string
		originalUrl    string
		id             int64
		input          *dynamodb.PutItemInput
		output         *dynamodb.PutItemOutput
		writeItemError error
		checkError     bool
	}{
		"WriteItem Happy Path": {
			shortUrl:    "b",
			originalUrl: "http://www.example.com",
			id:          int64(1),
			input: &dynamodb.PutItemInput{
				TableName: &tableName,
				Item: map[string]types.AttributeValue{
					ID:          &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", int64(1))},
					ShortURL:    &types.AttributeValueMemberS{Value: "b"},
					OriginalURL: &types.AttributeValueMemberS{Value: "http://www.example.com"},
				},
			},
			output: &dynamodb.PutItemOutput{}, // we don't care about this
		},
		"WriteItem Sad Path": {
			shortUrl:    "b",
			originalUrl: "http://www.example.com",
			id:          int64(1),
			input: &dynamodb.PutItemInput{
				TableName: &tableName,
				Item: map[string]types.AttributeValue{
					ID:          &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", int64(1))},
					ShortURL:    &types.AttributeValueMemberS{Value: "b"},
					OriginalURL: &types.AttributeValueMemberS{Value: "http://www.example.com"},
				},
			},
			writeItemError: errors.New("error"),
			checkError:     true,
		},
	}
	logger, _ := zap.NewProduction()

	for _, tc := range tests {
		m := &MockDynamoDBClient{}

		m.On("PutItem", context.Background(), tc.input).Return(tc.output, tc.writeItemError)

		db := &UrlDB{
			Logger:    logger,
			DBClient:  m,
			TableName: URLTable,
		}

		err := db.WriteItem(tc.id, tc.shortUrl, tc.originalUrl)

		if tc.checkError {
			assert.Error(t, err)
			return
		}

		m.AssertNumberOfCalls(t, "PutItem", 1)
	}
}
