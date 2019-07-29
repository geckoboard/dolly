.PHONY=build clean

build:  bin/deploy

clean:
	rm -rf bin

bin/dolly:
	mkdir -p bin
	env GOOS=linux GOARCH=arm GOARM=5 go build -o bin/dolly ./... 


release: bin/dolly
	cat bin/dolly | ssh pi@192.168.40.157  "tee dolly > /dev/null && chmod a+x dolly && sudo systemctl stop dolly && sudo mv dolly /usr/local/bin/dolly && sudo systemctl start dolly"
