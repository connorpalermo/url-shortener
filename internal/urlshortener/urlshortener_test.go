package urlshortener

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockDBProvider struct {
	mock.Mock
}

func (m *MockDBProvider) GetItemByPK(shortUrl string) (*dynamodb.GetItemOutput, error) {
	args := m.Called(shortUrl)
	return args.Get(0).(*dynamodb.GetItemOutput), args.Error(1)
}

func (m *MockDBProvider) WriteItem(id int64, shortUrl, originalUrl string) error {
	args := m.Called(id, shortUrl, originalUrl)
	return args.Error(0)
}

func (m *MockDBProvider) IncrementCounter() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockDBProvider) GetItemByNonPK(attributeName, attributeValue string) (*dynamodb.ScanOutput, error) {
	args := m.Called(attributeName, attributeValue)
	return args.Get(0).(*dynamodb.ScanOutput), args.Error(1)
}

func Test_ShortenURL(t *testing.T) {
	tests := map[string]struct {
		orignalURL  string
		scanOutput  *dynamodb.ScanOutput
		countValue  int64
		scanError   error
		writeError  error
		countError  error
		shortURL    string
		expectError bool
	}{
		"Happy Path URL shorten valid": {
			orignalURL: "http://www.example.com",
			scanOutput: &dynamodb.ScanOutput{
				Count: 0,
			},
			countValue: int64(1),
			shortURL:   "b",
		},
		"Happy Path URL shorten seen before": {
			orignalURL: "http://www.example.com",
			scanOutput: &dynamodb.ScanOutput{
				Count: 1,
				Items: []map[string]types.AttributeValue{
					{
						"original_url": &types.AttributeValueMemberS{Value: "https://example.com"},
						"short_url":    &types.AttributeValueMemberS{Value: "b"},
					},
				},
			},
			countValue: int64(1),
			shortURL:   "b",
		},
		"Sad Path Scan Error": {
			orignalURL:  "http://www.example.com",
			scanError:   errors.New("error"),
			expectError: true,
		},
		"Sad Path no short_url attribute": {
			orignalURL: "http://www.example.com",
			scanOutput: &dynamodb.ScanOutput{
				Count: 1,
				Items: []map[string]types.AttributeValue{
					{
						"original_url": &types.AttributeValueMemberS{Value: "https://example.com"},
						"invalid":      &types.AttributeValueMemberS{Value: "b"},
					},
				},
			},
			expectError: true,
		},
		"Sad Path IncrementCount error": {
			orignalURL: "http://www.example.com",
			scanOutput: &dynamodb.ScanOutput{
				Count: 0,
			},
			countError:  errors.New("error"),
			expectError: true,
		},
		"Sad Path WriteItem error": {
			orignalURL: "http://www.example.com",
			scanOutput: &dynamodb.ScanOutput{
				Count: 0,
			},
			countValue:  int64(1),
			writeError:  errors.New("error"),
			expectError: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			logger, err := zap.NewProduction()
			if err != nil {
				return
			}

			m := new(MockDBProvider)

			u := &UrlShortener{
				Logger:   logger,
				DBClient: m,
			}

			m.On("WriteItem", mock.Anything, mock.Anything, mock.Anything).Return(tc.writeError)
			m.On("IncrementCounter").Return(tc.countValue, tc.countError)

			m.On("GetItemByNonPK", mock.Anything, tc.orignalURL).Return(tc.scanOutput, tc.scanError)
			shortened, err := u.ShortenURL(tc.orignalURL)

			if tc.expectError {
				assert.Error(t, err)
				return
			}
			assert.Equal(t, shortened, tc.shortURL)
		})
	}
}

func Test_GetOriginalURL(t *testing.T) {
	tests := map[string]struct {
		orignalURL  string
		shortURL    string
		itemOutput  *dynamodb.GetItemOutput
		dbError     error
		expectError bool
	}{
		"Happy Path URL exists": {
			orignalURL: "http://www.example.com",
			shortURL:   "b",
			itemOutput: &dynamodb.GetItemOutput{
				Item: map[string]types.AttributeValue{
					"id":           &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", int64(1))},
					"short_url":    &types.AttributeValueMemberS{Value: "b"},
					"original_url": &types.AttributeValueMemberS{Value: "http://www.example.com"},
				},
			},
		},
		"Sad Path Get Item By PK error": {
			orignalURL:  "http://www.example.com",
			shortURL:    "b",
			dbError:     errors.New("error"),
			expectError: true,
		},
		"Sad Path No OriginalURL Attribute": {
			orignalURL: "http://www.example.com",
			shortURL:   "b",
			itemOutput: &dynamodb.GetItemOutput{
				Item: map[string]types.AttributeValue{
					"id":        &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", int64(1))},
					"short_url": &types.AttributeValueMemberS{Value: "b"},
				},
			},
			expectError: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			logger, err := zap.NewProduction()
			if err != nil {
				return
			}
			m := new(MockDBProvider)
			m.On("GetItemByPK", tc.shortURL).Return(tc.itemOutput, tc.dbError)
			u := &UrlShortener{
				Logger:   logger,
				DBClient: m,
			}

			original, err := u.GetOriginalURL(tc.shortURL)

			if tc.expectError {
				assert.Error(t, err)
				return
			}
			assert.Equal(t, original, tc.orignalURL)
		})
	}
}

func Test_NewClient(t *testing.T) {
	logger, _ := zap.NewProduction()
	u, _ := New(logger)
	u.Logger.Info("Successfully Initialized")
}
