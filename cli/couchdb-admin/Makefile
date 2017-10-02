GOARCH = amd64
PREFIX = "couchdb-admin"
VERSION?=?
COMMIT=$(shell git rev-parse HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

LDFLAGS = -ldflags "-X main.VERSION=${VERSION} -X main.COMMIT=${COMMIT} -X main.BRANCH=${BRANCH}"

all: clean default

default: 
	GOOS=linux GOARCH=${GOARCH} go build ${LDFLAGS} -o ${PREFIX}

linux: 
	GOOS=linux GOARCH=${GOARCH} go build ${LDFLAGS} -o ${PREFIX}-linux-${GOARCH}

darwin: 
	GOOS=darwin GOARCH=${GOARCH} go build ${LDFLAGS} -o ${PREFIX}-darwin-${GOARCH}

clean:
	-rm -f ${PREFIX}*

.PHONY: default linux darwin clean
