build:
	go build ./...
	go build github.com/birdayz/fancyfs/cmd/fancy
	go build github.com/birdayz/fancyfs/cmd/fancyd
install:
	go install ./...
test:
	go test -v -cover  ./...
lint:
	gometalinter -j4 --disable=interfacer --exclude="\bexported \w+ (\S*['.]*)([a-zA-Z'.*]*) should have comment or be unexported\b" --vendor ./...