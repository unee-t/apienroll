DEVUPJSON = '.profile |= "uneet-dev" \
		  |.stages.staging |= (.domain = "unit.dev.unee-t.com" | .zone = "dev.unee-t.com") \
		  | .actions[0].emails |= ["kai.hendry+unitdev@unee-t.com"] \
		  | .lambda.vpc.subnets |= [ "subnet-0e123bd457c082cff", "subnet-0ff046ccc4e3b6281", "subnet-0e123bd457c082cff" ] \
		  | .profile |= "uneet-dev" \
		  | .lambda.vpc.security_groups |= [ "sg-0b83472a34bc17400", "sg-0f4dadb564041855b" ]'

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
	jq '.profile |= "uneet-demo" |.stages.staging |= (.domain = "apienroll.demo.unee-t.com" | .zone = "demo.unee-t.com")' up.json.in > up.json
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

.PHONY: dev demo prod
