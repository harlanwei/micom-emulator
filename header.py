# Vian Chen <imvianchen@stu.pku.edu.cn>

import os
import json

dir = os.path.dirname(os.path.realpath(__file__))
refcodes = []
with open(f"{dir}/refcodes.json", "r") as f:
    obj = json.load(f)
    refcodes = obj["refcodes"]

content = f"""// This file is generated automatically.
// Do not write changes to this file.

#pragma once

#define MAX_CODE {len(refcodes)}

static const char *comm_desc[MAX_CODE] = {{
"""

for code in refcodes:
    content += f"\t\"{code[1]}\",\n"

content += "};\n"

with open(f"{dir}/micom/refcodes.h", "w+") as header:
    header.write(content)
