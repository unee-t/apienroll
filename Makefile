# This script is called by the `deploy.sh` file in (this folder)
# We use this to deploy the environments with Travis CI

# We create a function to simplify getting variables for aws parameter store.

define ssm
$(shell aws --profile $(TRAVIS_AWS_PROFILE) ssm get-parameters --names $1 --with-decryption --query Parameters[0].Value --output text)
endef

# We prepare variables for up in UPJSON and PRODUPJSON.
# These variables are coming from AWS Parameter Store
# - STAGE
# - DOMAIN
# - EMAIL_FOR_NOTIFICATION_APIENROLL
# - PRIVATE_SUBNET_1
# - PRIVATE_SUBNET_2
# - PRIVATE_SUBNET_3
# - LAMBDA_TO_RDS_SECURITY_GROUP

UPJSON = '.profile |= "$(TRAVIS_AWS_PROFILE)" \
		  |.stages.staging |= (.domain = "apienroll.$(call ssm,STAGE).$(call ssm,DOMAIN)" | .zone = "$(call ssm,STAGE).$(call ssm,DOMAIN)") \
		  | .actions[0].emails |= ["$(call ssm,EMAIL_FOR_NOTIFICATION_APIENROLL)"] \
		  | .lambda.vpc.subnets |= [ "$(call ssm,PRIVATE_SUBNET_1)", "$(call ssm,PRIVATE_SUBNET_2)", "$(call ssm,PRIVATE_SUBNET_3)" ] \
		  | .lambda.vpc.security_groups |= [ "$(call ssm,LAMBDA_TO_RDS_SECURITY_GROUP)" ]'

#UPJSON for Production

PRODUPJSON = '.profile |= "$(TRAVIS_AWS_PROFILE)" \
		  |.stages.staging |= (.domain = "apienroll.$(call ssm,DOMAIN)" | .zone = "$(call ssm,DOMAIN)") \
		  | .actions[0].emails |= ["$(call ssm,EMAIL_FOR_NOTIFICATION_APIENROLL)"] \
		  | .lambda.vpc.subnets |= [ "$(call ssm,PRIVATE_SUBNET_1)", "$(call ssm,PRIVATE_SUBNET_2)", "$(call ssm,PRIVATE_SUBNET_3)" ] \
		  | .lambda.vpc.security_groups |= [ "$(call ssm,LAMBDA_TO_RDS_SECURITY_GROUP)" ]'

dev:
	@echo $$AWS_ACCESS_KEY_ID
	jq $(UPJSON) up.json.in > up.json
	up

prod:
	@echo $$AWS_ACCESS_KEY_ID
	jq $(PRODUPJSON) up.json.in > up.json
	up

demo:
	@echo $$AWS_ACCESS_KEY_ID
	jq $(UPJSON) up.json.in > up.json
	up

testping:
	curl -i -H "Authorization: Bearer $(call ssm,API_ACCESS_TOKEN)" https://apienroll.$(call ssm,STAGE).$(call ssm,DOMAIN)
