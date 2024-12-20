#!/bin/bash

# Configuration
FUNCTION_NAME="urlShortenerLambda"
ROLE_NAME="urlShortenerRole"
TABLE_NAME="url-mapping"
S3_BUCKET="url-shortener-source"
ZIP_FILE="function.zip"
API_NAME="urlShortenerAPI"
REGION="us-east-1"

# Package Lambda function
echo "Packaging Lambda function..."
make clean
make zip

# Create S3 Bucket if it doesn't exist
echo "Creating S3 bucket..."
if [ "$REGION" = "us-east-1" ]; then
  aws s3api create-bucket --bucket $S3_BUCKET
else
  aws s3api create-bucket --bucket $S3_BUCKET --create-bucket-configuration LocationConstraint=$REGION --region $REGION
fi

# Upload to S3
echo "Uploading ZIP file to S3..."
aws s3 cp $ZIP_FILE s3://$S3_BUCKET/

# Create DynamoDB table
echo "Creating DynamoDB table..."
aws dynamodb create-table \
    --table-name $TABLE_NAME \
    --attribute-definitions AttributeName=short_url,AttributeType=S \
    --key-schema AttributeName=short_url,KeyType=HASH \
    --billing-mode PAY_PER_REQUEST \
    --region $REGION

# Create IAM Role for Lambda
echo "Creating IAM Role..."
ROLE_POLICY_DOCUMENT='{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}'
echo $ROLE_POLICY_DOCUMENT > trust-policy.json
aws iam create-role --role-name $ROLE_NAME --assume-role-policy-document file://trust-policy.json

# Attach permissions to role
echo "Attaching permissions to IAM Role..."
aws iam attach-role-policy --role-name $ROLE_NAME --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole
aws iam attach-role-policy --role-name $ROLE_NAME --policy-arn arn:aws:iam::aws:policy/AmazonDynamoDBFullAccess

# Wait for role propagation
echo "Waiting for IAM Role to propagate..."
sleep 10

# Deploy Lambda function
echo "Deploying Lambda function..."
ROLE_ARN=$(aws iam get-role --role-name $ROLE_NAME --query "Role.Arn" --output text)
aws lambda create-function \
    --function-name $FUNCTION_NAME \
    --runtime provided.al2 \
    --role $ROLE_ARN \
    --handler main \
    --code S3Bucket=$S3_BUCKET,S3Key=$ZIP_FILE \
    --region $REGION

# Create regional REST API
echo "Creating REST API..."
API_ID=$(aws apigateway create-rest-api \
    --name $API_NAME \
    --region $REGION \
    --endpoint-configuration types=REGIONAL \
    --query "id" --output text)

echo "Created API with ID: $API_ID"

# Get the root resource ID
ROOT_RESOURCE_ID=$(aws apigateway get-resources \
    --rest-api-id $API_ID \
    --region $REGION \
    --query "items[?path=='/'].id" --output text)

echo "Root resource ID: $ROOT_RESOURCE_ID"

# Create POST /shorten resource
echo "Creating POST /shorten resource..."
POST_RESOURCE_ID=$(aws apigateway create-resource \
    --rest-api-id $API_ID \
    --parent-id $ROOT_RESOURCE_ID \
    --path-part "shorten" \
    --region $REGION \
    --query "id" --output text)

echo "POST /shorten resource ID: $POST_RESOURCE_ID"

# Create GET /{shortUrl} resource
echo "Creating GET /{shortUrl} resource..."
GET_RESOURCE_ID=$(aws apigateway create-resource \
    --rest-api-id $API_ID \
    --parent-id $ROOT_RESOURCE_ID \
    --path-part "{shortUrl}" \
    --region $REGION \
    --query "id" --output text)

echo "GET /{shortUrl} resource ID: $GET_RESOURCE_ID"

# Add POST /shorten method to API Gateway
echo "Adding POST /shorten method to API Gateway..."
aws apigateway put-method \
    --rest-api-id $API_ID \
    --resource-id $POST_RESOURCE_ID \
    --http-method POST \
    --authorization-type NONE \
    --region $REGION

# Add GET /{shortUrl} method to API Gateway
echo "Adding GET /{shortUrl} method to API Gateway..."
aws apigateway put-method \
    --rest-api-id $API_ID \
    --resource-id $GET_RESOURCE_ID \
    --http-method GET \
    --authorization-type NONE \
    --region $REGION

# Add method response for POST /shorten
echo "Adding 200 OK response for POST /shorten..."
aws apigateway put-method-response \
    --rest-api-id $API_ID \
    --resource-id $POST_RESOURCE_ID \
    --http-method POST \
    --status-code 200 \
    --region $REGION

# Add method response for GET /{shortUrl}
echo "Adding 200 OK response for GET /{shortUrl}..."
aws apigateway put-method-response \
    --rest-api-id $API_ID \
    --resource-id $GET_RESOURCE_ID \
    --http-method GET \
    --status-code 200 \
    --region $REGION

# Add a small delay before creating integrations
echo "Waiting for methods to propagate before creating integrations..."
sleep 5

# Create Lambda integration for POST /shorten route
echo "Creating Lambda integration for POST /shorten route..."
POST_INTEGRATION_ID=$(aws apigateway put-integration \
    --rest-api-id $API_ID \
    --resource-id $POST_RESOURCE_ID \
    --http-method POST \
    --integration-http-method POST \
    --type AWS_PROXY \
    --uri arn:aws:apigateway:$REGION:lambda:path/2015-03-31/functions/$(aws lambda get-function --function-name $FUNCTION_NAME --query "Configuration.FunctionArn" --output text)/invocations \
    --region $REGION \
    --query "id" --output text)

echo "Created Lambda Integration with ID: $POST_INTEGRATION_ID"

# Create Lambda integration for GET /{shortUrl} route
echo "Creating Lambda integration for GET /{shortUrl} route..."
GET_INTEGRATION_ID=$(aws apigateway put-integration \
    --rest-api-id $API_ID \
    --resource-id $GET_RESOURCE_ID \
    --http-method GET \
    --integration-http-method GET \
    --type AWS_PROXY \
    --uri arn:aws:apigateway:$REGION:lambda:path/2015-03-31/functions/$(aws lambda get-function --function-name $FUNCTION_NAME --query "Configuration.FunctionArn" --output text)/invocations \
    --region $REGION \
    --query "id" --output text)

echo "Created Lambda Integration with ID: $GET_INTEGRATION_ID"

# Add permission for API Gateway to invoke Lambda
echo "Granting API Gateway permission to invoke Lambda..."
API_GATEWAY_ARN="arn:aws:execute-api:$REGION:$(aws sts get-caller-identity --query "Account" --output text):$API_ID/*/*/*"

aws lambda add-permission \
  --function-name $FUNCTION_NAME \
  --statement-id apigateway-invoke-permission \
  --action lambda:InvokeFunction \
  --principal apigateway.amazonaws.com \
  --source-arn $API_GATEWAY_ARN \
  --region $REGION

echo "Permission granted for API Gateway to invoke Lambda."

# Create the stage 'prod' before deploying
echo "Creating stage 'prod'..."
aws apigateway create-deployment \
    --rest-api-id $API_ID \
    --stage-name prod \
    --region $REGION

# Deploy API Gateway
echo "Deploying API Gateway..."
aws apigateway create-deployment --rest-api-id $API_ID --region $REGION --stage-name prod

# Output API URL
API_URL="https://$API_ID.execute-api.$REGION.amazonaws.com/prod"
echo "API Gateway is deployed. You can use the following URL for access: $API_URL"

echo "Done! API Gateway deployed, Lambda, DynamoDB, and S3 setup complete."
