DEVUPJSON = '.profile |= "uneet-dev" \
		  |.stages.staging |= (.domain = "apienroll.dev.unee-t.com" | .zone = "dev.unee-t.com") \
		  | .actions[0].emails |= ["kai.hendry+apienrolldev@unee-t.com"] \
		  | .lambda.vpc.subnets |= [ "subnet-0e123bd457c082cff", "subnet-0ff046ccc4e3b6281", "subnet-0e123bd457c082cff" ] \
		  | .lambda.vpc.security_groups |= [ "sg-0b83472a34bc17400", "sg-0f4dadb564041855b" ]'


DEMOUPJSON = '.profile |= "uneet-demo" \
		  |.stages.staging |= (.domain = "apienroll.demo.unee-t.com" | .zone = "demo.unee-t.com") \
		  | .actions[0].emails |= ["kai.hendry+apienrolldemo@unee-t.com"] \
		  | .lambda.vpc.subnets |= [ "subnet-0bdef9ce0d0e2f596", "subnet-091e5c7d98cd80c0d", "subnet-0fbf1eb8af1ca56e3" ] \
		  | .lambda.vpc.security_groups |= [ "sg-6f66d316" ]'


NAME=apienroll
REPO=uneet/$(NAME)

all:
	go build

dev:
	@echo $$AWS_ACCESS_KEY_ID
	jq $(DEVUPJSON) up.json.in > up.json
	up

demo:
	@echo $$AWS_ACCESS_KEY_ID
	jq $(DEMOUPJSON) up.json.in > up.json
	up

prod:
	@echo $$AWS_ACCESS_KEY_ID
	jq '.profile |= "uneet-prod" |.stages.staging |= (.domain = "apienroll.unee-t.com" | .zone = "unee-t.com")' up.json.in > up.json
	up

clean:
	rm -f apienroll gin-bin

build:
	docker build -t $(REPO) --build-arg COMMIT=$(shell git describe --always) .

start:
	docker run -d --name $(NAME) -p 9000:9000 $(REPO)

stop:
	docker stop $(NAME)
	docker rm $(NAME)

sh:
	docker exec -it $(NAME) /bin/sh


testping:
	curl -i -H "Authorization: Bearer $(shell aws --profile uneet-demo ssm get-parameters --names API_ACCESS_TOKEN --with-decryption --query Parameters[0].Value --output text)" https://apienroll.demo.unee-t.com
