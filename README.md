# Overview:

This codebase allows us to deploy the Unee-T APIs to add and manage users.

This codbase uses AWS Lambdas and relies on AWS Aurora's capability to call lambdas directly from Database envent (CALL and TRIGGER).

# Pre-Requisite:

- This is intended to be deployed on AWS.
- We use Travis CI for automated deployment.
- One of the dependencies for this repo is maintained on the [unee-t/env codebase](https://github.com/unee-t/env)

The following variables MUST be declared in order for this to work as intended:

## AWS variables:

These should be decleared in the AWS Parameter Store for this environment.

- DOMAIN
- INSTALLATION_ID
- STAGE
- EMAIL_FOR_NOTIFICATION_APIENROLL
- PRIVATE_SUBNET_1
- PRIVATE_SUBNET_2
- PRIVATE_SUBNET_3
- LAMBDA_TO_RDS_SECURITY_GROUP
- API_ACCESS_TOKEN

Make sure to check the AWS variables needed by the unee-t/env codebase in the [pre-requisite described in the README file]](https://github.com/unee-t/env#pre-requisite).

## Travic CI variables:

These should be declared as Settings in Travis CI for this Repository.

### For all environments:
 - DOCKER_USERNAME
 - DOCKER_PASSWORD
 - AWS_DEFAULT_REGION

### For dev environment:
 - AWS_ACCOUNT_USER_ID_DEV
 - AWS_ACCOUNT_SECRET_DEV
 - AWS_PROFILE_DEV

### For Demo environment:
 - AWS_ACCOUNT_USER_ID_DEMO
 - AWS_ACCOUNT_SECRET_DEMO
 - AWS_PROFILE_DEMO

### For Prod environment:
 - AWS_ACCOUNT_USER_ID_PROD
 - AWS_ACCOUNT_SECRET_PROD
 - AWS_PROFILE_PROD

# Deployment:

Deployment is done automatically with Travis CI:
- For the DEV environment: each time there is a change in the `master` repo for this codebase.
- For the PROD and DEMO environment: each time we do a tag release for this repo.

# Maintenance:

To get the latest version of the go modules we need. 
You can run
`go get -u`

See the [documentation on go modules](https://blog.golang.org/using-go-modules) for more details.

# Debugging:

Because we are relying heavily on AWS, it impossible to debug locally.

## Before deployment:

Tests are done on Travis CI before deployment

## When the Lambdas have been deployed to AWS:

Everything is running on AWS this is where you'll find useful information to debug

# Logs:

## Before deployment:

Check the Travis CI log for the build job for this repository each time you 
- push a new PR
- Merge a PR into the Master Branch
- Create a new Tag Release

## When the Lambdas have been deployed to AWS:

Logs are available in the AWS Cloudwatch for your environment under `log groups` >> `/aws/lambda/apienroll`

## Check if things were correctly deployed:

You can check if a build has been completed successfully in a given environment by checking the AWS ECS >> Task Definition 

**TODO: add more information on how we can find more info on the date of the latest Task that was run (date, build, etc...)**

# Common issues:

## API Access token issues:
 
If you change the AWS Secret for the API_ACCESS_TOKEN then you MUST redeploy all the codebase
- frontend
- apienroll
- unit
- invite
- lambda2sqs
- etc...

# More information:

The API is also documented in our `postman@unee-t.com` Postman account.
