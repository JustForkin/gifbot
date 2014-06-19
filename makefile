all:
	go get ./web && go build -o gifweb ./web
	go get ./bot && go build -o gifbot ./bot

bot:
	go get ./bot && go build -o gifbot ./bot

web:
	go get ./web && go build -o gifweb ./web
