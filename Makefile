.PHONY: build run tidy docker

build:
	go build -o manager-services ./app

run:
	go run ./app --db=services.db --address=:8090 --debug

tidy:
	go mod tidy

docker:
	docker build -t manager-services .

clean:
	rm -f manager-services services.db
