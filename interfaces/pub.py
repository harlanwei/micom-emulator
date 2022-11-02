# Vian Chen <imvianchen@stu.pku.edu.cn>

import json
import re
import os
import sys
from typing import Any, Callable, List
from functools import partial

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.dirname(SCRIPT_DIR))
from interfaces import impl

REF_COMMAND = 0

log_err = partial(print, file=sys.stderr)

# Reference: https://stackoverflow.com/questions/1628949/to-find-first-n-prime-numbers-in-python
def get_first_n_primes(n: int) -> List[int]:
    def isPrime(n: int) -> bool:
        return re.match(r'^1?$|^(11+?)\1+$', '1' * n) == None
    primes = []
    step = 100
    upper_bound = step + 2
    while len(primes) < n:
        primes = [i for i in range(upper_bound - step, upper_bound) if isPrime(i)]
        upper_bound += step
    return primes[:n]

refcodes = []
dir = os.path.dirname(os.path.realpath(__file__))
with open(f"{dir}/../refcodes.json", "r") as f:
    obj = json.load(f)
    refcodes = obj["refcodes"]
command_repr = get_first_n_primes(len(refcodes))

def call(cb: Callable[[int], Any], command: str) -> None:
    for ind, ref in enumerate(refcodes):
        if re.compile(ref[REF_COMMAND]).match(command):
            cb(command_repr[ind])
            exit(0)
    
    log_err(f"Unknown command: {command}")

if __name__ == '__main__':
    if len(sys.argv) != 3:
        print('Error: incorrect number of args')
        exit(-1)

    interface = sys.argv[1]
    command = sys.argv[2].lower()

    if interface == 'ioctl':
        call(impl.ioctl, command)
    elif interface == 'mmio':
        call(impl.mmio, command)
    elif interface == 'procfs':
        call(impl.procfs, command)
    elif interface == 'netlink':
        call(impl.netlink, command)
    else:
        log_err(f"No such interface: {interface}")
    
