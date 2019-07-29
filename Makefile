
clean:
	rm -rf bin

bin/release:
	mkdir -p bin
	env GOOS=linux GOARCH=arm GOARM=5 go build -o bin/release ./... 
