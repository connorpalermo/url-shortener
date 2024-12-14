package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	chiadapter "github.com/awslabs/aws-lambda-go-api-proxy/chi"
	"github.com/connorpalermo/url-shortener/internal/endpoint"
	"github.com/connorpalermo/url-shortener/internal/persistence"
	"github.com/connorpalermo/url-shortener/internal/router"
	"github.com/connorpalermo/url-shortener/internal/urlshortener"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	db, err := persistence.New(logger)
	if err != nil {
		logger.Error("failed to initialize db client")
		return
	}

	u := &urlshortener.UrlShortener{
		Logger:   logger,
		DBClient: db,
	}

	mux := router.New(&endpoint.Handler{
		Logger:               logger,
		UrlShortenerProvider: u,
	})

	chiLambda := chiadapter.New(mux)

	// Start Lambda handler with ProxyWithContext
	lambda.Start(func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		logger.Info("Raw Request", zap.Any("request", request))
		return chiLambda.ProxyWithContext(ctx, request)
	})
}
