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

Deployment is done automatically with Tracis CI:
- For the DEV environment: each time there is a change in the `master` repo for this codebase
- For the PROD and DEMO environment: each time we do a tag release for this repo.

# More information:

The API is also documented in our `postman@unee-t.com` Postman account.
