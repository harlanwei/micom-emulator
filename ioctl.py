import fcntl
import os

if __name__ == '__main__':
    print('hello')
    with open("/dev/micom", "wb") as fd:
        fcntl.ioctl(fd, 0)