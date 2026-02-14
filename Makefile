IMAGE   := latex-renderer
NAME    := latex-test
PORT    := 8080
API_KEY := test123
PROFILE := iamadmin-general
REGION  := us-east-1

# --- Auto tag ---
TAG := $(shell git rev-parse --short HEAD)

.PHONY: build run stop restart logs prod-url clean deploy update destroy

# ---- Local ----

build:
	docker build --provenance=false -t $(IMAGE):$(TAG) .

run:
	docker run -d --name $(NAME) -p $(PORT):8080 -e API_KEY=$(API_KEY) $(IMAGE):$(TAG)

stop:
	docker stop $(NAME) && docker rm $(NAME)

restart: stop run

logs:
	docker logs -f $(NAME)

clean:
	-docker stop $(NAME) 2>/dev/null
	-docker rm $(NAME) 2>/dev/null
	docker rmi $(IMAGE)
	-rm -f output.html output.pdf

prod-url:
	cd infra && terraform output -raw api_url

# ---- AWS Deploy ----

deploy:
		@echo "== 1/4 Terraform init =="
	cd infra && terraform init

	@echo "== 2/4 Creating ECR repository =="
	cd infra && terraform apply -target=aws_ecr_repository.this \
		-var="api_key=$(API_KEY)" \
		-var="image_uri=placeholder" \
		-auto-approve

	@echo "== 3/4 Build & push image =="
	$(eval ECR_URL := $(shell cd infra && terraform output -raw ecr_repository_url))

	docker build --provenance=false -t $(IMAGE):$(TAG) .
	aws ecr get-login-password --region $(REGION) --profile $(PROFILE) \
		| docker login --username AWS --password-stdin $(ECR_URL)

	docker tag $(IMAGE):$(TAG) $(ECR_URL):$(TAG)
	docker push $(ECR_URL):$(TAG)

	@echo "== 4/4 Deploying Lambda + API Gateway =="
	cd infra && terraform apply \
		-var="api_key=$(API_KEY)" \
		-var="image_uri=$(ECR_URL):$(TAG)" \
		-auto-approve

	@echo ""
	@echo "Deployed with image tag: $(TAG)"
	@cd infra && terraform output api_url

update:
	$(eval ECR_URL := $(shell cd infra && terraform output -raw ecr_repository_url))

	docker build --provenance=false -t $(IMAGE):$(TAG) .
	aws ecr get-login-password --region $(REGION) --profile $(PROFILE) \
		| docker login --username AWS --password-stdin $(ECR_URL)

	docker tag $(IMAGE):$(TAG) $(ECR_URL):$(TAG)
	docker push $(ECR_URL):$(TAG)

	aws lambda update-function-code \
		--function-name $(IMAGE) \
		--image-uri $(ECR_URL):$(TAG) \
		--profile $(PROFILE) \
		--no-cli-pager

	@aws lambda wait function-updated \
		--function-name $(IMAGE) \
		--profile $(PROFILE)

	@echo "Updated Lambda to tag $(TAG)"

destroy:
	cd infra && terraform destroy \
		-var="api_key=$(API_KEY)" -var="image_uri=placeholder" -auto-approve
	@echo "All AWS resources destroyed."
