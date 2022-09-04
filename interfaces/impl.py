# Vian Chen <imvianchen@stu.pku.edu.cn>

import fcntl
import mmap
import sys
import os

# ======================= IOCTL Linux helpers =======================

_IOC_WRITE = 1

_IOC_NRBITS = 8
_IOC_TYPEBITS = 8
_IOC_SIZEBITS = 14

_IOC_NRSHIFT = 0
_IOC_TYPESHIFT = _IOC_NRSHIFT + _IOC_NRBITS
_IOC_SIZESHIFT = _IOC_TYPESHIFT + _IOC_TYPEBITS
_IOC_DIRSHIFT = _IOC_SIZESHIFT + _IOC_SIZEBITS

def _IOC(dir, type, nr, size):
    return dir  << _IOC_DIRSHIFT  | \
           type << _IOC_TYPESHIFT | \
           nr   << _IOC_NRSHIFT   | \
           size << _IOC_SIZESHIFT

def _IOW(type, nr, size):
    return _IOC(_IOC_WRITE, type, nr, size)

def MICOM_IOW(nr: int) -> int:
    return _IOW(0x15, nr, 0)

# ===================================================================

def ioctl(code: int):
    with open("/dev/micom", "wb") as f:
        fcntl.ioctl(f, MICOM_IOW(code))

def mmio(code: int):
    fd = os.open("/dev/micom", os.O_RDWR | os.O_SYNC)
    mm = mmap.mmap(fd, 0x1000)
    mm[0] = code
    mm.close()
    os.close(fd)

def procfs(code: int):
    with open("/proc/micom", "wb") as file:
        file.write(code)

def netlink(code: int = 0):
    print(
        "Error: Netlink is already deprecated in Linux.",
        file=sys.stderr
    )
