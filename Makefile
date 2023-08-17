.PHONY : snapshot

PROFILE=$(shell aws configure list-profiles | grep personal || echo "default")
ACCOUNT=$(shell aws sts get-caller-identity --profile $(PROFILE) | jq -r '.Account')
GITHUB_TOKEN=$(shell aws ssm get-parameter --profile $(PROFILE) --name /github_token --with-decryption | jq -r '.Parameter.Value')

# replace with alias for remote server
REMOTE_ALIAS=proto
REMOTE_ARCH=$(shell ssh $(REMOTE_ALIAS) "uname -m")

# problem to execute
PROBLEM=problem-11
PORT=4269

snapshot:
	ACCOUNT=$(ACCOUNT) PROFILE=$(PROFILE) goreleaser build --snapshot --rm-dist

release:
	aws ecr get-login-password  --profile $(PROFILE) --region ca-central-1 | docker login --username AWS --password-stdin $(ACCOUNT).dkr.ecr.ca-central-1.amazonaws.com
	GITHUB_TOKEN=$(GITHUB_TOKEN) ACCOUNT=$(ACCOUNT) PROFILE=$(PROFILE) goreleaser release --rm-dist

proto: snapshot
	scp dist/protohackers_linux_arm64/protohackers $(REMOTE_ALIAS):~/protohackers;\
