# micom-emulator

This project creates an emulation environment for [this](https://github.com/rapid7/metasploit-framework/blob/master/modules/post/android/local/koffee.rb) Metasploit module. It composes of an emulated `micom` driver, an emulated `micomd` daemon and a userspace program `watchdog` that shows if an action is successfully triggered.

Tested on Debian 11 with kernel 5.10.127.

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

## To-dos

- [ ] Improve watchdog interface to better emulate a car HU
