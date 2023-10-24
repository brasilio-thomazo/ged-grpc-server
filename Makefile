gen:
	protoc \
	--proto_path=proto \
	--go_out=pb \
	--go_opt=paths=source_relative \
	--go-grpc_out=pb \
	--go-grpc_opt=paths=source_relative \
	proto/*.proto

clean:
	rm -f pb/*.pb.go

run:
	go run main.go

build:
	go build -v -o hermes .

docker-build:
	docker build -t devoptimus/ged-grpc-server --target=server .

docker-push:
	docker push devoptimus/ged-grpc-server

docker-up: docker-down
	docker compose up -d --build

docker-down:
	docker compose down