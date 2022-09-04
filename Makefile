all: driver clangd-conf
	cd micom && make load
	cd ..
	rm -f watchdog && ln -s $(shell pwd)/watchdog-client/watchdog watchdog

header:
	python3 ./header.py

driver:
	cd micom && make all
	cd ..

watchdog: watchdog-client/client.go
	cd watchdog-client && go build
	cd ..

clangd-conf:
	cd micom; make clean; bear -- make

clean:
	rm -f micom/compile_commands.json watchdog interface
	rm -rf interfaces/__pycache__
	cd micom && make clean