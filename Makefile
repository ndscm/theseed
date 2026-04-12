.PHONY: all bootstrap format tidy test

all: format tidy test

bootstrap:
	./bootstrap.sh

format:
	./format.sh

test: bootstrap
	./test.sh

tidy: bootstrap
	./tidy.sh
