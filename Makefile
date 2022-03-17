DIST=./dist
BIN=rivendell
OS_MAC=darwin
ARCH_MAC=amd64
OS_LINUX=linux
ARCH_LINUX=amd64
OS_M1=darwin
ARCH_M1=arm64

build:
	go build

install:
	go install

clean:
	rm -rf ${BIN} ${DIST} ${GOPATH}/bin/${BIN}

dist: clean build .dist-prepare .dist-mac .dist-linux .dist-m1

test: .test-project .test-utils .test-kubernetes

.dist-prepare:
	rm -rf ${DIST}
	mkdir -p ${DIST}

.dist-mac:
	CGO_ENABLED=0 GOOS=${OS_MAC} GOARCH=${ARCH_MAC} go build -o ${DIST}/${BIN} && \
	cd ${DIST} && \
	tar czf ${BIN}-`../${BIN} version`-${OS_MAC}-${ARCH_MAC}.tar.gz ${BIN} && \
	rm ${BIN} && \
	cd ..

.dist-m1:
	CGO_ENABLED=0 GOOS=${OS_M1} GOARCH=${ARCH_M1} go build -o ${DIST}/${BIN} && \
	cd ${DIST} && \
	tar czf ${BIN}-`../${BIN} version`-${OS_M1}-${ARCH_M1}.tar.gz ${BIN} && \
	rm ${BIN} && \
	cd ..

.dist-linux:
	CGO_ENABLED=0 GOOS=${OS_LINUX} GOARCH=${ARCH_LINUX} go build -o ${DIST}/${BIN} && \
	cd ${DIST} && \
	tar czf ${BIN}-`../${BIN} version`-${OS_LINUX}-${ARCH_LINUX}.tar.gz ${BIN} && \
	rm ${BIN} && \
	cd ..

.test-project:
	go test ./project -v

.test-utils:
	go test ./utils -v

.test-kubernetes:
	go test ./kubernetes -v
