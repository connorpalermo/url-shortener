# URL Shortener Deployment Script

This script automates the deployment process for a URL shortener service built in Golang. The service exposes two main endpoints:

- `POST /shorten`: Accepts a URL to shorten and returns the shortened version.
  - **Request Body**:
    ```json
    {
      "originalUrl": "<longUrl>"
    }
    ```
  - **Response**:
    - Returns a shortened URL that can be used to access the original URL.

- `GET /{shortUrl}`: Retrieves the original URL associated with the provided shortened URL.
  - **Request**:
    - URL path parameter: `{shortUrl}` (the shortened URL identifier).
  - **Response**:
    - Returns the original URL that corresponds to the provided shortened URL

The script performs the following tasks:

1. **Package the Lambda function**: Uses `make` to clean and package the Golang Lambda function into a ZIP file.
2. **Create AWS Resources**:
   - Creates an S3 bucket to store the Lambda function ZIP file.
   - Creates a DynamoDB table (`url-mapping`) to store mappings between shortened URLs and their original counterparts.
   - Creates an IAM role and attaches necessary policies to allow Lambda to access DynamoDB and execute with basic Lambda permissions.
   - Deploys the Lambda function to AWS.
3. **Set up API Gateway**:
   - Creates a regional REST API in API Gateway.
   - Defines two routes: `POST /shorten` and `GET /{shortUrl}`, linking them to the Lambda function.
   - Adds method responses and integrates the Lambda function with the API Gateway.
   - Grants API Gateway the necessary permissions to invoke the Lambda function.
   - Deploys the API to the `prod` stage.

## Prerequisites

- AWS CLI configured with the necessary permissions to create and manage IAM roles, Lambda functions, API Gateway, S3, and DynamoDB resources.
- Golang `make` utility to package the Lambda function.
- A working Lambda function in Golang, packaged into a ZIP file (`function.zip`).

## Script Breakdown

### Configuration
You can modify the following variables at the beginning of the script to fit your environment:
- `FUNCTION_NAME`: Name of the Lambda function (default: `urlShortenerLambda`).
- `ROLE_NAME`: Name of the IAM role for Lambda (default: `urlShortenerRole`).
- `TABLE_NAME`: Name of the DynamoDB table (default: `url-mapping`).
- `S3_BUCKET`: Name of the S3 bucket to store the Lambda code (default: `url-shortener-source`).
- `ZIP_FILE`: Name of the Lambda function ZIP file (default: `function.zip`).
- `API_NAME`: Name of the API Gateway (default: `urlShortenerAPI`).
- `REGION`: AWS region (default: `us-east-1`).

### Steps

1. **Packaging Lambda**: The Lambda function code is cleaned and packaged into a ZIP file (`function.zip`) using `make`.
2. **Creating S3 Bucket**: Creates an S3 bucket to store the Lambda code. If the region is `us-east-1`, the bucket is created without a region specification.
3. **Creating DynamoDB Table**: Creates a DynamoDB table (`url-mapping`) with `short_url` as the primary key.
4. **Creating IAM Role**: Creates an IAM role for Lambda with permissions to execute and interact with DynamoDB.
5. **Deploying Lambda**: Deploys the packaged Lambda function to AWS using the IAM role created earlier.
6. **Setting up API Gateway**: Creates a regional REST API with two resources:
   - `POST /shorten`: Shortens a URL.
   - `GET /{shortUrl}`: Resolves a shortened URL to its original.
7. **Integrating Lambda with API Gateway**: Configures API Gateway to forward requests to the Lambda function, both for `POST` and `GET` methods.
8. **Permissions**: Grants API Gateway the permission to invoke the Lambda function.
9. **Deploy API Gateway**: Deploys the API to the `prod` stage, making the API live and accessible.

### Output
Once the script is executed, the following will be displayed:
- API Gateway URL: The URL to access the deployed URL shortener API.

Example output:

API Gateway is deployed. You can use the following URL for access: https://<api-id>.execute-api.us-east-1.amazonaws.com/prod

## Usage

Run this script from your terminal with the necessary AWS permissions. The script will automate the entire deployment process and output the URL of the deployed API once complete.

```bash
$ chmod +x ./deploy-url-shortener.sh
$ ./deploy-url-shortener.sh