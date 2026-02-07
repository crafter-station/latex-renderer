IMAGE   := latex-renderer
NAME    := latex-test
PORT    := 8080
API_KEY := test123
PROFILE := iamadmin-general
REGION  := us-east-1

.PHONY: build run stop restart logs test test-remote clean deploy update destroy

# ---- Local ----

build:
	docker build --provenance=false -t $(IMAGE) .

run:
	docker run -d --name $(NAME) -p $(PORT):8080 -e API_KEY=$(API_KEY) $(IMAGE)

stop:
	docker stop $(NAME) && docker rm $(NAME)

restart: stop run

logs:
	docker logs -f $(NAME)

test:
	@curl -s -X POST http://localhost:$(PORT)/render \
		-H "Authorization: Bearer $(API_KEY)" \
		-H "Content-Type: text/plain" \
		--data-binary @test.tex \
		-o output.html -w "\nHTTP %{http_code} - saved to output.html\n"
	@cat output.html

clean:
	-docker stop $(NAME) 2>/dev/null
	-docker rm $(NAME) 2>/dev/null
	docker rmi $(IMAGE)
	-rm -f output.html

test-remote:
	$(eval API_URL := $(shell cd infra && terraform output -raw api_url))
	@curl -s -X POST $(API_URL)render \
		-H "Authorization: Bearer $(API_KEY)" \
		-H "Content-Type: text/plain" \
		--data-binary @test.tex \
		-o output.html -w "\nHTTP %{http_code} - saved to output.html\n"
	@cat output.html

# ---- AWS Deploy ----

deploy:
	@echo "== 1/4 Terraform init =="
	cd infra && terraform init
	@echo "== 2/4 Creating ECR repository =="
	cd infra && terraform apply -target=aws_ecr_repository.this \
		-var="api_key=$(API_KEY)" -var="image_uri=placeholder" -auto-approve
	@echo "== 3/4 Pushing image to ECR =="
	$(eval ECR_URL := $(shell cd infra && terraform output -raw ecr_repository_url))
	aws ecr get-login-password --region $(REGION) --profile $(PROFILE) \
		| docker login --username AWS --password-stdin $(ECR_URL)
	docker tag $(IMAGE):latest $(ECR_URL):latest
	docker push $(ECR_URL):latest
	@echo "== 4/4 Deploying Lambda + API Gateway =="
	cd infra && terraform apply \
		-var="api_key=$(API_KEY)" -var="image_uri=$(ECR_URL):latest" -auto-approve
	@echo ""
	@echo "Deployed! API URL:"
	@cd infra && terraform output api_url

update:
	$(eval ECR_URL := $(shell cd infra && terraform output -raw ecr_repository_url))
	docker build --provenance=false -t $(IMAGE) .
	aws ecr get-login-password --region $(REGION) --profile $(PROFILE) \
		| docker login --username AWS --password-stdin $(ECR_URL)
	docker tag $(IMAGE):latest $(ECR_URL):latest
	docker push $(ECR_URL):latest
	aws lambda update-function-code --function-name $(IMAGE) \
		--image-uri $(ECR_URL):latest --profile $(PROFILE) --no-cli-pager
	@echo "Updated! Waiting for Lambda to be ready..."
	@aws lambda wait function-updated --function-name $(IMAGE) --profile $(PROFILE)
	@echo "Done."

destroy:
	cd infra && terraform destroy \
		-var="api_key=$(API_KEY)" -var="image_uri=placeholder" -auto-approve
	@echo "All AWS resources destroyed."
