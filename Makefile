all: driver
	cd micom && make load
	cd ..

header:
	python3 ./header.py

driver: header
	cd micom && make all
	cd ..

clean:
	cd micom && rm refcodes.h -f && make clean