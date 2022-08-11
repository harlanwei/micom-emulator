all: driver
	cd micom && make load
	cd ..
	ln -s $(shell pwd)/watchdog-client/watchdog watchdog

header:
	python3 ./header.py

driver: header
	cd micom && make all
	cd ..

watchdog:
	cd watchdog && go build

clean:
	rm include/refcodes.h watchdog -f && cd micom && make clean