# This code will need the following variables
# - Installation ID (AWS parameter `INSTALLATION_ID`)
#	- Stage (AWS parameter `STAGE`) 

# TODO 
# Make sure we tackle the problematic assumptions:
# Code assumes that there is a user with userId = 32 in the installation.
# Code assumes that userApiKey for user 32 has a certain value in this installation.

curl -i -X POST \
  http://localhost:3000 \
  -H "Authorization: Bearer $(aws --profile ${INSTALLATION_ID}-${STAGE} ssm get-parameters --names API_ACCESS_TOKEN --with-decryption --query Parameters[0].Value --output text)" \
  -H 'Cache-Control: no-cache' \
  -H 'Content-Type: application/json' \
  -d '{
	"userId": "32",
	"userApiKey": "kwmJ6OT6tzEdmctCrT3uEt7kMnLEqlzckW4eaTvX"
}'
