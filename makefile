# Docker build (prod)
build_feed_docker_image_prod:
	docker build -f docker/production/Dockerfile.feed -t feed-service-prod:latest .

build_user_docker_image_prod:
	docker build -f docker/production/Dockerfile.user -t user-service-prod:latest .

build_gateway_docker_image_prod:
	docker build -f docker/production/Dockerfile.gateway -t api-gateway-prod:latest .

build_all_service_docker_image_prod: build_feed_docker_image_prod build_user_docker_image_prod build_gateway_docker_image_prod

push_gateway_image_to_hub:
	docker tag api-gateway-prod:latest baby831109/api-gateway-prod:latest
	docker push baby831109/api-gateway-prod:latest  

push_user_image_to_hub:
	docker tag user-service-prod:latest baby831109/user-service-prod:latest
	docker push baby831109/user-service-prod:latest  

push_feed_image_to_hub:
	docker tag feed-service-prod:latest baby831109/feed-service-prod:latest
	docker push baby831109/feed-service-prod:latest  

push_all_service_to_docker_hub: push_feed_image_to_hub push_gateway_image_to_hub push_user_image_to_hub


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

# Kubernates deploy 
deploy_user_service:
	kubectl delete secret user-secret --ignore-not-found
	kubectl create secret generic user-secret \
		--from-env-file=user-service/.env.prod

	kubectl apply -f k8s/production/user-service-deployment.yaml

deploy_feed_service:
	kubectl delete secret feed-secret --ignore-not-found
	kubectl create secret generic feed-secret \
		--from-env-file=feed-service/.env.prod

	kubectl apply -f k8s/production/feed-service-deployment.yaml

deploy_api_gateway:
	kubectl delete secret gateway-secret --ignore-not-found
	kubectl create secret generic gateway-secret \
		--from-env-file=api-gateway/.env.prod

	kubectl apply -f k8s/production/api-gateway-deployment.yaml

deploy_prod: deploy_user_service deploy_feed_service deploy_api_gateway

rollout_user_service:
	kubectl rollout restart deployment user-service

rollout_feed_service:
	kubectl rollout restart deployment feed-service

rollout_api_gateway:
	kubectl rollout restart deployment api-gateway

rollout_all_service: rollout_api_gateway rollout_feed_service rollout_user_service

#minikube load images
load_user_image:
	minikube image load baby831109/user-service-prod:latest

load_feed_image:
	minikube image load baby831109/feed-service-prod:latest
	
load_gateway_image:
	minikube image load baby831109/api-gateway-prod:latest

minikube_load_all_images: load_gateway_image load_feed_image load_user_image