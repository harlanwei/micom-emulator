# Vian Chen <imvianchen@stu.pku.edu.cn>

import subprocess
import os
import json
import re
from typing import List

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
    

dir = os.path.dirname(os.path.realpath(__file__))
refcodes = []
with open(f"{dir}/refcodes.json", "r") as f:
    obj = json.load(f)
    refcodes = obj["refcodes"]

chdr_content = f"""// This file is generated automatically.
// Do not write changes to this file.

#pragma once

#define N_COMMANDS {len(refcodes) + 1}

__attribute__((__used__))
static const char *comm_desc[{len(refcodes) + 1}] = {{
    "update_eventfd",
"""

gohdr_content = f"""// This file is generated automatically.
// Do not write changes to this file.

package main

const (
"""

command_repr = get_first_n_primes(len(refcodes))

for ind, code in enumerate(refcodes):
    chdr_content += f"\t\"{code[1]}\",\n"
    gohdr_content += f"\t{code[1].upper()} = {command_repr[ind]}\n"

chdr_content += f"""}};

__attribute__((__used__))
static const int desc_ind[{len(refcodes) + 1}] = {{
    0, // for eventfd
"""
gohdr_content += """)

var CodeEventMap = map[uint64]string{
"""

for ind, code in enumerate(refcodes):
    chdr_content += f"\t{command_repr[ind]},\n"
    gohdr_content += f"\t{command_repr[ind]}: \"{code[1].lower()}\",\n"

chdr_content += "};"
gohdr_content += "}"

chdr_path = f"{dir}/micom/refcodes.h"
os.makedirs(os.path.dirname(chdr_path), exist_ok=True)
with open(chdr_path, "w+", encoding="utf-8") as header:
    header.write(chdr_content)

gohdr_path = f"{dir}/watchdog-client/constants.go"
os.makedirs(os.path.dirname(gohdr_path), exist_ok=True)
with open(gohdr_path, "w+", encoding="utf-8") as header:
    header.write(gohdr_content)
subprocess.check_output(["go", "fmt", "watchdog-client/constants.go"])
