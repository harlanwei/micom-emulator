LKM = micom

obj-m += ${LKM}.o
$(LKM)-objs += main.o uevent.o proc.o mmio.o

all:
	make -C /lib/modules/$(shell uname -r)/build M=$(PWD) modules

clean:
	make -C /lib/modules/$(shell uname -r)/build M=$(PWD) clean

unload:
	sudo rmmod $(LKM).ko || true

load: unload all
	sudo insmod $(LKM).ko
	sudo dmesg | tail
