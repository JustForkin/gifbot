all:
	rm gifweb && go get ./web && go build -o gifweb ./web
	rm gifbot && go get ./bot && go build -o gifbot ./bot

bot:
	rm gifbot && go get ./bot && go build -o gifbot ./bot

web:
	rm gifweb && go get ./web && go build -o gifweb ./web
