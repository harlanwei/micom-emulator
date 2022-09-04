# Vian Chen <imvianchen@stu.pku.edu.cn>

import subprocess
import os
import json

dir = os.path.dirname(os.path.realpath(__file__))
refcodes = []
with open(f"{dir}/refcodes.json", "r") as f:
    obj = json.load(f)
    refcodes = obj["refcodes"]

chdr_content = f"""// This file is generated automatically.
// Do not write changes to this file.

#pragma once

#define MAX_CODE {len(refcodes)}

__attribute__((__used__))
static const char *comm_desc[{len(refcodes) + 1}] = {{
    "", // for eventfd
"""

gohdr_content = f"""// This file is generated automatically.
// Do not write changes to this file.

package main

const (
"""

for ind, code in enumerate(refcodes):
    chdr_content += f"\t\"{code[1]}\",\n"
    gohdr_content += f"\t{code[1].upper()} = {ind+1}\n"

chdr_content += "};\n"
gohdr_content += ")"

chdr_path = f"{dir}/micom/refcodes.h"
os.makedirs(os.path.dirname(chdr_path), exist_ok=True)
with open(chdr_path, "w+", encoding="utf-8") as header:
    header.write(chdr_content)

gohdr_path = f"{dir}/watchdog-client/constants.go"
os.makedirs(os.path.dirname(gohdr_path), exist_ok=True)
with open(gohdr_path, "w+", encoding="utf-8") as header:
    header.write(gohdr_content)
subprocess.check_output(["go", "fmt", "watchdog-client/constants.go"])