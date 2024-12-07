package urlshortener

import (
	"errors"
	"sync"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/connorpalermo/url-shortener/constant/logkey"
	urlDB "github.com/connorpalermo/url-shortener/internal/persistence"
	"go.uber.org/zap"
)

type (
	UrlShortener struct {
		Logger   *zap.Logger
		Mu       sync.Mutex
		DBClient URLDBProvider
	}

	UrlShortenerProvider interface {
		ShortenURL(url string) (string, error)
		GetOriginalURL(shortened string) (string, error)
	}

	URLDBProvider interface {
		GetItemByPK(shortUrl string) (*dynamodb.GetItemOutput, error)
		WriteItem(id int64, shortUrl, originalUrl string) error
		IncrementCounter() (int64, error)
		GetItemByNonPK(attributeName, attributeValue string) (*dynamodb.ScanOutput, error)
	}
)

const (
	Base62Chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	OriginalURL = "original_url"
	ShortURL    = "short_url"
)

func New(logger *zap.Logger) (*UrlShortener, error) {
	db, err := urlDB.New(logger)
	if err != nil {
		return nil, err
	}

	return &UrlShortener{
		Logger:   logger,
		DBClient: db,
	}, nil
}

func (u *UrlShortener) ShortenURL(url string) (string, error) {
	u.Mu.Lock()
	defer u.Mu.Unlock()

	scan, err := u.DBClient.GetItemByNonPK(OriginalURL, url)
	if err != nil {
		return "", err
	}

	if scan.Count > 0 {
		// we have already seen this URL
		item := scan.Items[0]
		var attributes map[string]string
		err = attributevalue.UnmarshalMap(item, &attributes)
		if err != nil {
			return "", err
		}

		shortURL, ok := attributes[ShortURL]
		if !ok {
			return "", errors.New("ShortURL attribute missing")
		}

		return shortURL, nil
	}

	u.Logger.Info("shortening original URL: ", zap.String(logkey.OriginalURL, url))

	id, err := u.DBClient.IncrementCounter()
	if err != nil {
		return "", err
	}
	shortened := encodeBase62(id)

	err = u.DBClient.WriteItem(id, shortened, OriginalURL)
	if err != nil {
		return "", err
	}
	u.Logger.Info("generated shortened URL: ", zap.String(logkey.ShortenedURL, shortened))

	return shortened, nil
}

func encodeBase62(id int64) string {
	result := ""
	for id > 0 {
		rem := id % 62
		result = string(Base62Chars[rem]) + result
		id /= 62
	}
	return result
}

func (u *UrlShortener) GetOriginalURL(shortened string) (string, error) {
	u.Mu.Lock()
	defer u.Mu.Unlock()

	u.Logger.Info("getting original URL from shortened URL: ", zap.String(logkey.ShortenedURL, shortened))

	original, err := u.DBClient.GetItemByPK(shortened)
	if err != nil {
		return "", err
	}

	var attributes map[string]string
	err = attributevalue.UnmarshalMap(original.Item, &attributes)
	if err != nil {
		return "", err
	}

	originalURL, ok := attributes[OriginalURL]
	if !ok {
		return "", errors.New("OriginalURL attribute missing")
	}

	u.Logger.Info("retrieved original URL: ", zap.String(logkey.OriginalURL, originalURL))

	return originalURL, nil
}
