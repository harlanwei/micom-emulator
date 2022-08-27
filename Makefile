all: driver
	cd micom && make load
	cd ..
	rm -f watchdog && ln -s $(shell pwd)/watchdog-client/watchdog watchdog

header:
	python3 ./header.py

driver: header
	cd micom && make all
	cd ..

watchdog: watchdog-client/client.go
	cd watchdog-client && go build
	cd ..

clean:
	rm -f micom/refcodes.h watchdog interface
	rm -rf interfaces/__pycache__
	cd micom && make clean