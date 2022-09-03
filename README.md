# micom-emulator

This project creates an emulation environment for [this](https://github.com/rapid7/metasploit-framework/blob/master/modules/post/android/local/koffee.rb) Metasploit module. It composes of an emulated `micom` driver, an emulated `micomd` daemon and a userspace program `watchdog` that shows if an action is successfully triggered.

Tested on:

- Debian 11 with kernel 5.10.127
- Ubuntu 22.04.1 with kernel 5.15.0-46-generic

Note that you can't use this on WSL by default, since WSL (even WSL2, which does have a standalone virtualized kernel) does not allow loading custom kernel modules. To use with WSL, you would need to compile your own kernel.

## Compile & Load

Simply run `make all` under this directory. Do make sure you have root access or at least can use `sudo`.

## Usage

```bash
# compile & load as described above
make all

# execute userspace watchdog
./watchdog

# in a new bash
./micomd -c inject [command]
```

All available commands are in the `refcodes.json` file. For example, `./micomd -c inject 0112 f0` toggles maximum radio volume.

If the commands are successfully injected, watchdog should print out messages identifying each command.

### With Metasploit

Below is an example of establishing a Metasploit session with your local machine.

```bash
msf > use auxiliary/scanner/ssh/ssh_login
msf auxiliary(ssh_login) > set rhosts 127.0.0.1
msf auxiliary(ssh_login) > set username <your username>
msf auxiliary(ssh_login) > set password <your password>
msf auxiliary(ssh_login) > exploit

# At this point you should have a session with id `1`. To
# make sure, you can list all your sessions:
msf > sessions -l

msf > use post/android/local/koffee
msf post(koffee) > set session 1 # or your session id
msf post(koffee) > set micomd <path to micomd>
msf post(koffee) > exploit # or other actions
```

## References

[1] Costantino, G., & Matteucci, I. (2020). KOFFEE-Kia OFFensivE Exploit. Istituto di Informatica e Telematica, Tech. Rep.
