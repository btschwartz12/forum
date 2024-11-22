sqlc:
	cd repo/db && sqlc generate

swagger:
	swag init --output server/api/swagger -g server/api/swagger/main.go

forum: sqlc swagger
	CGO_ENABLED=0 go build -o forum main.go

run-server: forum
	godotenv -f .env ./forum \
		--port 8000 \
		--var-dir var \
		--public

clean:
	rm -f forum
