
# Manage our application builder.
.PHONY: build-prod run test clean

build-prod:
	@go build -tags prod embed -o ./bin/app cmd/main.go

prod-test:
	@docker-compose -f docker-compose.yml down
	@docker-compose -f docker-compose.yml --profile prod-test up

run:
	@go build -tags dev -o ./bin/app cmd/main.go
	@./bin/app

test:
	@go test ./...

clean:
	@go clean
air:
	@${GOPATH}/bin/air -c .air.toml

# Manage our infrastructure using docker compose
.PHONY: start, stop, bounce

infra-start:
	docker-compose -f docker-compose.yml up
infra-stop:
	docker-compose -f docker-compose.yml down

infra-bounce: infra-stop infra-start