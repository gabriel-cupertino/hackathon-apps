up:
	docker compose up -d
down:
	docker compose down
build:
	docker compose build --no-cache
all:
	docker compose down
	docker compose build --no-cache
	docker compose up -d
test:
	pip install -r requirements-dev.txt -q && pytest test_app.py -v
