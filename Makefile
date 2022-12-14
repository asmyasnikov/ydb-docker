YDB_VERSION = 22.4.44
CLI_VERSION = 2.1.0
IMAGE=ydb-platform/yandex-docker-local-ydb
TAG=$(YDB_VERSION)

artifacts:
	rm -rf artifacts && mkdir -p artifacts/{bin,lib};

download-ydbd: artifacts
	cd artifacts && wget https://binaries.ydb.tech/release/${YDB_VERSION}/ydbd-${YDB_VERSION}-linux-amd64.tar.gz && tar --strip-components=1 -xvzf ydbd-${YDB_VERSION}-linux-amd64.tar.gz && rm -rf ydbd-${YDB_VERSION}-linux-amd64.tar.gz;

download-cli: artifacts
	cd artifacts/bin && wget https://storage.yandexcloud.net/yandexcloud-ydb/release/${CLI_VERSION}/linux/amd64/ydb;

download: download-ydbd download-cli
	# nothing to do for this target

clone:
	# clone some branch/tag

compile: artifacts clone
	# compile from cloned sources

docker:
	@docker build -t $(IMAGE):$(YDB_VERSION) .

docker_push:
	@docker push $(IMAGE):$(TAG)
