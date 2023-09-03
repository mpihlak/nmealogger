.PHONY: build

build:
	docker build -t nmealogger-build .
	docker rm nmealogger-build-c >/dev/null 2>&1 || true
	docker create --name=nmealogger-build-c nmealogger-build
	docker cp nmealogger-build-c:/nmealogger_1.0-1_all.deb .
	docker rm nmealogger-build-c
