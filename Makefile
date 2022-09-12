all: driver clangd-conf
	cd micom && make load
	cd ..
	rm -f watchdog && ln -s $(shell pwd)/watchdog-client/watchdog watchdog
	echo "key value" > scene.tmp

header:
	python3 ./header.py

driver:
	cd micom && make all
	cd ..

watchdog: watchdog-client/*.go
	cd watchdog-client && go build
	cd ..

clangd-conf:
	cd micom; make clean; LLVM=1 bear -- make

clean:
	rm -rf micom/.cache micom/compile_commands.json
	rm -rf interfaces/__pycache__
	rm -f watchdog 
	rm -f scene.tmp interface
	cd micom && make clean