# We create a function to simplify getting variables for aws parameter store.
# The variable TRAVIS_AWS_PROFILE is set when .travis.yml runs

define ssm
$(shell aws --profile $(TRAVIS_AWS_PROFILE) ssm get-parameters --names $1 --with-decryption --query Parameters[0].Value --output text)
endef

# The other variables needed up in UPJSON and PRODUPJSON are set when `source ./aws.env` runs
# - STAGE
# - DOMAIN
# - EMAIL_FOR_NOTIFICATION_APIENROLL
# - PRIVATE_SUBNET_1
# - PRIVATE_SUBNET_2
# - PRIVATE_SUBNET_3
# - LAMBDA_TO_RDS_SECURITY_GROUP
# These variables can be edited in the AWS parameter store for the environment

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

# We have everything, we can run up now.

dev:
	@echo $$AWS_ACCESS_KEY_ID
	# We replace the relevant variable in the up.json file
	# We use the template defined in up.json.in for that
	jq $(UPJSON) up.json.in > up.json
	up

demo:
	@echo $$AWS_ACCESS_KEY_ID
	# We replace the relevant variable in the up.json file
	# We use the template defined in up.json.in for that
	jq $(UPJSON) up.json.in > up.json
	up

prod:
	@echo $$AWS_ACCESS_KEY_ID
	# We replace the relevant variable in the up.json file
	# We use the template defined in up.json.in for that
	jq $(PRODUPJSON) up.json.in > up.json
	up

testping:
	curl -i -H "Authorization: Bearer $(API_ACCESS_TOKEN)" https://apienroll.$(STAGE).$(DOMAIN)
