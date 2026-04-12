.PHONY: all bootstrap sanitize test

all: sanitize test

bootstrap:
	./bootstrap.sh

sanitize: bootstrap
	./sanitize.sh

test: bootstrap
	./test.sh
