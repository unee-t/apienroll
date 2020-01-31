#!/bin/bash

# This script is created to deploy apienroll 
# Variable is the environment like dev, demo or prod
# To run this script, run this command: ./deploy.sh [STAGE] 
# where STAGE is 
#   - `dev` for the DEV environment
#   - `demo` for the DEMO environment
#   - `prod` for the PROD environment
#
##Step 1: Setup AWS CLI Profile
# This is in case there is no aws cli profile
# in that case, the aws profile needs to be created from scratch.
# This happens when:
#	- We are doing a travis CI deployment.
#	  We rely on the Travis CI settings that have been called when the
#	  .travis.yml script is called.
#	- The user has not configured his machine properly.
#
# Step 1: Setup AWS CLI
if ! aws configure --profile $TRAVIS_AWS_PROFILE list
then
    # We tell the user about the issue
	echo Profile $TRAVIS_AWS_PROFILE does not exist >&2

	if ! test "$TRAVIS_AWS_ACCESS_KEY_ID"
	then
        # We tell the user about the issue
		echo Missing $TRAVIS_AWS_ACCESS_KEY_ID >&2
		exit 1
	fi

	echo Attempting to setup one from the environment >&2
	aws configure set profile.${TRAVIS_AWS_PROFILE}.aws_access_key_id $TRAVIS_AWS_ACCESS_KEY_ID
	aws configure set profile.${TRAVIS_AWS_PROFILE}.aws_secret_access_key $TRAVIS_AWS_SECRET_ACCESS_KEY
	aws configure set profile.${TRAVIS_AWS_PROFILE}.region $TRAVIS_AWS_DEFAULT_REGION

	if ! aws configure --profile $TRAVIS_AWS_PROFILE list
	then
		echo Profile $TRAVIS_AWS_PROFILE does not exist >&2
		exit 1
	fi

fi

# Step 2: Run Make
    make $1