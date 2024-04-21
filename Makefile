test:
	go test -race
	go test -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html