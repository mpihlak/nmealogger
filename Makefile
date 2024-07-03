.PHONY: build

build:
	docker build -t nmealogger-build .
	docker rm nmealogger-build-c >/dev/null 2>&1 || true
	docker create --name=nmealogger-build-c nmealogger-build
	docker cp nmealogger-build-c:/nmealogger_`head -1 debian/changelog | cut -f2 -d'(' | cut -f1 -d')'`_all.deb .
	docker rm nmealogger-build-c
