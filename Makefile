all: driver
	cd micom && make load
	cd ..

header:
	python3 ./header.py

driver: header
	cd micom && make all
	cd ..

watchdog:
	cd watchdog && go build

clean:
	rm include/refcodes.h -f && cd micom && make clean