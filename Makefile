down\:%:
	docker compose down $*

up\:%:
	docker compose up -d $*

stresstest\:%: up\:mockorigin up\:%
	sudo docker compose run --rm -e TARGET_HOST=$* k6 run /benchmarks/stresstest.ts
	make down:$*
	make down:mockorigin 2> /dev/null

benchmark\:%: up\:mockorigin up\:%
	sudo docker compose run --rm -e TARGET_HOST=$* k6 run /benchmarks/benchmark.ts
	make down:$*
	make down:mockorigin 2> /dev/null

