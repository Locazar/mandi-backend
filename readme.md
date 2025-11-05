# Ecommerce Project

### 1. Clone the Repository
Clone the mandi-backend repository to your local system:
```bash
git clone https://github.com/Rohit221990/mandi-backend.git
cd mandi-backend
```
### 2. Install Dependencies
Install the required dependencies using either the provided Makefile command or Go's built-in module management:
```bash
# Using Makefile
make deps
# OR using Go
go mod tidy
```
### 3. Configure Environment Variables
details provided at the end of the file
### 4. Make Swagger Files (For Swagger API Documentation)
```bash
make swag
```
# To Run The Application
```bash
make run
```
### To See The API Documentation
1. visit [swagger] ***http://localhost:3000/swagger/index.html***

# To Test The Application
### 1. Generate Mock Files
```bash
make mockgen
```
### 2. Run The Test Functions
```bash
make test
```

# Set up Environment Variables
Set up the necessary environment variables in a .env file at the project's root directory. Below are the variables required:
```.env
### PostgreSQL database details
DB_NAME="your database name"
DB_USER="your database user name"
DB_PASSWORD="your database owner password"
DB_PORT="your database running port number"
### JWT
ADMIN_AUTH_KEY="secret code for signing admin JWT token"
USER_AUTH_KEY="secret code for signing user JWT token"
### Twilio
AUTH_TOKEN="your Twilio authentication token"
ACCOUNT_SID="your Twilio account SID"
SERVICE_SID="your Twilio messaging service SID"
### Razorpay
RAZOR_PAY_KEY="your Razorpay API test key"
RAZOR_PAY_SECRET="your Razorpay API test secret key"
### Stripe
STRIPE_SECRET="your Stripe account secret key"
STRIPE_PUBLISH_KEY="your Stripe account publish key"
STRIPE_WEBHOOK="your Stripe account webhook key"
### Google Auth
GOAUTH_CLIENT_ID="your Google auth client ID"
GOAUTH_CLIENT_SECRET="your Google auth secret key"
GOAUTH_CALL_BACK_URL="your registered callback URL for Google Auth"
### AWS S3 Service
AWS_ACCESS_KEY_ID="your aws access key id"
AWS_SECRET_ACCESS_KEY="your AWS secret access key"
AWS_REGION="your AWS region"
AWS_BUCKET_NAME="your AWS s3 bucket name"
```
