# Docker build (dev)
build_feed_docker_image_dev:
	docker build -f docker/development/Dockerfile.feed -t feed-service-dev:latest .

build_user_docker_image_dev:
	docker build -f docker/development/Dockerfile.user -t user-service-dev:latest .

build_gateway_docker_image_dev:
	docker build -f docker/development/Dockerfile.gateway -t api-gateway-dev:latest .

build_all_service_docker_image_dev: build_feed_docker_image_dev build_user_docker_image_dev build_gateway_docker_image_dev


# Docker compose
run_docker_compose_locally:
	docker compose --project-directory . -f docker/development/compose.yml up --build

clean_docker_compose_locally:
	docker compose --project-directory . -f docker/development/compose.yml down --rmi all --volumes
