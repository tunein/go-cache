SHELL := /bin/bash
GOENV = GOPRIVATE=github.com/tunein AWS_SDK_LOAD_CONFIG=true FAIL_DATADOG_INIT_SILENTLY=true
GOCMD = $(GOENV) go
GOLINT = ${GOENV} golint

AWS_DEFAULT_REGION ?= us-west-2
AWS_REGION ?= $(AWS_DEFAULT_REGION)
AWS_PROFILE ?= $(AWS_DEFAULT_PROFILE)

STREAMING_ECR_ID=658265200425
STREAMING_ECR_REGION=us-west-2
CONTAINER_REPO=builder-base
STREAMING_BUILD_CONTAINER_NAME?=streaming-build-container
STREAMING_BUILD_CONTAINER_VERSION?=0.54

RUN_CONTAINER = ${STREAMING_ECR_ID}.dkr.ecr.${STREAMING_ECR_REGION}.amazonaws.com/${CONTAINER_REPO}:${STREAMING_BUILD_CONTAINER_NAME}-${STREAMING_BUILD_CONTAINER_VERSION}

# if CI isn't set, force it to false. All TC builds should have it set to true
CI ?= false

ifeq ($(CI),true)
    AWS_DOCKER_ARGS = -e AWS_ACCESS_KEY_ID=$(AWS_ACCESS_KEY_ID) \
                      -e AWS_SECRET_ACCESS_KEY=$(AWS_SECRET_ACCESS_KEY) \
                      -e AWS_SESSION_TOKEN=$(AWS_SESSION_TOKEN) \
                      -e AWS_REGION=$(AWS_REGION)
    ECR_AUTHENTICATE_CMD = echo "skip ecr authentication, as it happens via github action"
    EXTRA_ARGS = -v ${HOME}/.gitconfig:/root/.gitconfig -v ${SSH_AUTH_SOCK}:${SSH_AUTH_SOCK} -e SSH_AUTH_SOCK=${SSH_AUTH_SOCK} --add-host=host.docker.internal:host-gateway
    GITCONFIG = chown root:root ~/.ssh/config && sed -i 's|/home/ubuntu|/root|g' ~/.ssh/config
    PRE_BUILD_COMMAND = printf '[safe]\n    directory = /go-cache' >> ~/.gitconfig && printf 'Host *\nStrictHostKeyChecking no\nCheckHostIP no\nTCPKeepAlive yes\nServerAliveInterval 30\nServerAliveCountMax 180\nVerifyHostKeyDNS yes\nUpdateHostKeys yes' >> ~/.ssh/config
else
    AWS_DOCKER_ARGS = -v $(HOME)/.aws:/root/.aws \
                      -e AWS_PROFILE=$(AWS_PROFILE) \
                      -e AWS_DEFAULT_PROFILE=$(AWS_DEFAULT_PROFILE) \
                      -e AWS_REGION=$(AWS_REGION)
    ECR_AUTHENTICATE_CMD = aws --profile=${AWS_PROFILE} ecr get-login-password --region ${AWS_REGION} | docker login --username AWS --password-stdin ${STREAMING_ECR_ID}.dkr.ecr.${STREAMING_ECR_REGION}.amazonaws.com
    EXTRA_ARGS = 
    GITCONFIG = git config --global url.\"git@github.com:tunein\".insteadOf \"https://github.com/tunein\"
    PRE_BUILD_COMMAND = echo
endif

DOCKERAWSCMD = docker run --rm \
                          -w /go-cache \
                          -v $$HOME/.ssh:/root/.ssh \
                          -v `pwd`:/go-cache \
                          $(AWS_DOCKER_ARGS) \
                          $(EXTRA_ARGS) \
                          --entrypoint /bin/bash \
                          $(RUN_CONTAINER) -c

all:
	@echo "making all..."
	@make update
	@make test
	@echo "make all complete"
.PHONY: all

authenticate:
	saml2aws login
.PHONY: authenticate

update:
	${GOCMD} mod tidy
.PHONY: update

ecr.authenticate:
	${ECR_AUTHENTICATE_CMD}
.PHONY: ec2.authenticate

coverage:
	@echo "running go-cache coverage..."
	@${GOCOVERAGE} --dir=. --output=coverage/inspector.coverage github.com/tunein/go-cache
	@${GOCMD} tool cover -html=coverage/inspector.coverage
	@echo "completed running inspector coverage."
.PHONY: coverage

test-setup:
	@echo "getting gotestrunner"
	rm -rf ../tmpsc
	git clone --no-checkout --depth=1 --no-tags https://github.com/tunein/go-common.git ../tmpsc
	cd ../tmpsc && git restore --staged scripts/gotestrunner && git checkout scripts/gotestrunner
	$(eval GOTESTRUNNER := "../tmpsc/scripts/gotestrunner")
.PHONY: test-setup

test: test-setup
	@echo "executing all tests..."
	FAIL_DATADOG_INIT_SILENTLY=true ${GOENV} ${GOTESTRUNNER}
	@echo "completed all tests."
.PHONY: test

%.test.docker:
	$(eval EXTRA_ARGS = ${EXTRA_ARGS}
	${DOCKERAWSCMD} "${GITCONFIG} && make $*.test"

test.docker:
	$(eval EXTRA_ARGS = ${EXTRA_ARGS})
	${PRE_BUILD_COMMAND}
	${DOCKERAWSCMD} "${GITCONFIG} && make test"

%.docker:
	$(eval EXTRA_ARGS = ${EXTRA_ARGS} -e CI=${CI})
	${PRE_BUILD_COMMAND}
	${DOCKERAWSCMD} "${GITCONFIG} && make $*"
.PHONY: %.docker
