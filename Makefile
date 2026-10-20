up:      docker compose up -d --build
down:    docker compose down
build:   go build ./...
run:     go run ./cmd/runner -scenario $(SCENARIO)
logs:    docker compose logs -f