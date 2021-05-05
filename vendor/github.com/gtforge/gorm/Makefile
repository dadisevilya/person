deps:
	curl -L https://s3.amazonaws.com/downloads.wercker.com/cli/stable/darwin_amd64/wercker -o /usr/local/bin/wercker
	chmod u+x /usr/local/bin/wercker

build: clean
	wercker build

build.local: clean
	wercker build --docker-local

clean:
	rm -rf .wercker

postgres.up:
	docker-compose up -d postgres

postgres.down:
	docker-compose down

postgres.restart: postgres.down postgres.up
	sleep 3

postgres.test: postgres.restart
	GORM_DIALECT=postgres go test -race ./...

test.all:
	docker-compose up -d
	sleep 3
	./test_all.sh