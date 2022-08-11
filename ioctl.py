# Vian Chen <imvianchen@stu.pku.edu.cn>

import fcntl
import json
import os
import sys
import re

REF_COMMAND = 0

# ======================= IOCTL Linux helpers =======================

_IOC_WRITE = 1

# constant for linux portability
_IOC_NRBITS = 8
_IOC_TYPEBITS = 8

# architecture specific
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

def IOW(nr: int) -> int:
    return _IOW(0x15, nr, 0)

# ===================================================================


if len(sys.argv) != 2:
    print('Error: incorrect number of args')
    exit(-1)

dir = os.path.dirname(os.path.realpath(__file__))
refcodes = []
with open(f"{dir}/refcodes.json", "r") as f:
    obj = json.load(f)
    refcodes = obj["refcodes"]

command = sys.argv[1].lower()
for ind, code in enumerate(refcodes):
    if re.compile(code[REF_COMMAND]).match(command):
        with open("/dev/micom", "wb") as fd:
            fcntl.ioctl(fd, IOW(ind + 1))
            exit(0)

print(f"Unknown command: {command}", file=sys.stderr)
