NAME=awsping
EXEC=${NAME}
BUILD_DIR=build
BUILD_OS="windows darwin freebsd linux"
BUILD_ARCH="amd64 386"
BUILD_DIR=build
SRC_CMD=cmd/awsping/main.go

build:
	go build -o ${EXEC} ${SRC_CMD}

# make run ARGS="-h"
run:
	go run cmd/awsping/main.go $(ARGS)

test:
	@go test -cover .

release: buildall
	git tag `grep "Version" utils.go | grep -o -E '[0-9]\.[0-9]\.[0-9]{1,2}'`
	git push --tags origin master

clean:
	@rm -f ${EXEC}
	@rm -f ${BUILD_DIR}/*
	@go clean

buildall: clean
	@mkdir -p ${BUILD_DIR}
	@for os in "${BUILD_OS}" ; do \
		for arch in "${BUILD_ARCH}" ; do \
			if [ $$os != "darwin" ] || [ $$arch != "386" ]; then \
				if [ $$os != "windows" ]; then \
					EXEC_FILE=${EXEC};\
				else\
					EXEC_FILE=${EXEC}.exe;\
				fi;\
				echo " * build $$os for $$arch"; \
				GOOS=$$os GOARCH=$$arch go build -ldflags "-s" -o ${BUILD_DIR}/$$EXEC_FILE ${SRC_CMD}; \
				cd ${BUILD_DIR}; \
				tar czf ${EXEC}.$$os.$$arch.tgz $$EXEC_FILE; \
				rm -f $$EXEC_FILE; \
				cd -  > /dev/null ; \
			fi;\
		done \
	done
	

docker:
	docker build -t awsping .

docker-run: docker
	docker run awsping -verbose 2 -repeats 2
