OUT = "bin"

all: build
build:
	go build -o $(OUT)/orb .
run:
	go run .
clean:
	rm $(OUT)/orb

test:
	go test -v .
