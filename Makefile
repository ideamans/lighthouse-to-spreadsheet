include .env
export $(shell sed 's/=.*//' .env)

lighthouse-to-spreadsheet: main.go
	go build -o lighthouse-to-spreadsheet main.go

.PHONY: clean
clean:
	rm -f lighthouse-to-spreadsheet

.PHONY: debug
debug: main.go
	go run main.go

.PHONY: install
install: lighthouse-to-spreadsheet
	sudo cp lighthouse-to-spreadsheet /usr/local/bin
