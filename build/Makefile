
all:
	@go install -tags=embed ../cmd/zipper/
	go build -tags=embed -o ./bin/zipper ../cmd/zipper/

run:
	@./start.sh

stop:
	@./stop.sh

clear:
	@rm -rf ./datadir

.PHONY: all run stop clear
