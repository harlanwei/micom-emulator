all: driver watchdog
	cd micom && make load
	cd ..

header:
	python3 ./header.py

driver: header
	cd micom && make all
	cd ..

watchdog:
	gcc -o watchdog watchdog.c -O3 -Wall

clean:
	rm include/refcodes.h watchdog -f && cd micom && make clean