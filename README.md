The API is documented in our `postman@unee-t.com` Postman account

# Local development

Required variables for database connectivity and authentication:

* MYSQL_HOST
* MYSQL_USER
* MYSQL_PASSWORD
* API_ACCESS_TOKEN

Can be overridden with local variables.

	while read -r; do export "$REPLY"; done < .env

# Docker image

[uneet/apienroll](https://hub.docker.com/r/uneet/apienroll/), as used in https://github.com/unee-t/bugzilla-customisation/blob/master/docker-compose.yml
