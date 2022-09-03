# Vian Chen <imvianchen@stu.pku.edu.cn>
# Generates an extension config for C/C++ on VSCode

import subprocess as sp
import re

kernel_version = sp.check_output(["uname", "-r"], encoding="utf-8").strip()
gcc_version_output = sp.check_output(["gcc", "--version"], encoding="utf-8").split('\n')[0]
gcc_version = re.search("gcc \(.*\) ([0-9]+)\.", gcc_version_output).group(1)

conf = f"""{{
    "configurations": [
        {{
            "name": "Linux",
            "includePath": [
                "${{workspaceFolder}}/**",
                "/usr/include",
                "/usr/local/include",
                "/usr/src/linux-headers-{kernel_version}/include",
                "/usr/src/linux-headers-{kernel_version}/arch/x86/include/**",
                "/usr/lib/gcc/x86_64-linux-gnu/{gcc_version}/include"
            ],
            "defines": [
                "__GNUC__",
                "__KERNEL__",
                "MODULE"
            ],
            "compilerPath": "/usr/bin/gcc",
            "cStandard": "gnu89",
            "cppStandard": "gnu++14",
            "intelliSenseMode": "linux-gcc-x64"
        }}
    ],
    "version": 4
}}"""

with open(".vscode/c_cpp_properties.json", "w") as f:
    f.write(conf)
