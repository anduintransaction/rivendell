install:
	go install

test: .test-project .test-utils .test-kubernetes

build: test
	goreleaser build --snapshot --rm-dist

.test-project:
	go test ./project -v

.test-utils:
	go test ./utils -v

.test-kubernetes:
	go test ./kubernetes -v
