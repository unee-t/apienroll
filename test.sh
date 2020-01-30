# Code on Line 9 IS WRONG: we should NOT use "ins-dev" here but the correct value based on
# - Installation ID (AWS parameter `INSTALLATION_ID`)
#	- Stage (AWS parameter `STAGE`) 
# The value for "ins-dev" is also declared in Travis Setting
# it is then passed on as the variable `TRAVIS_PROFILE`.

# Problematic assumptions:
# Code in line 15 assumes that there is a user with userId = 32 in the installation
# Code in line 16 assumes that userApiKey for user 32 has a certain value in this installation

curl -i -X POST \
  http://localhost:3000 \
  -H "Authorization: Bearer $(aws --profile ins-dev ssm get-parameters --names API_ACCESS_TOKEN --with-decryption --query Parameters[0].Value --output text)" \
  -H 'Cache-Control: no-cache' \
  -H 'Content-Type: application/json' \
  -d '{
	"userId": "32",
	"userApiKey": "kwmJ6OT6tzEdmctCrT3uEt7kMnLEqlzckW4eaTvX"
}'
