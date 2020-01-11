curl -i -X POST \
  http://localhost:3000 \
  -H "Authorization: Bearer $(aws --profile ins-dev ssm get-parameters --names API_ACCESS_TOKEN --with-decryption --query Parameters[0].Value --output text)" \
  -H 'Cache-Control: no-cache' \
  -H 'Content-Type: application/json' \
  -d '{
	"userId": "32",
	"userApiKey": "kwmJ6OT6tzEdmctCrT3uEt7kMnLEqlzckW4eaTvX"
}'
