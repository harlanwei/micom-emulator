# micom-emulator

This project creates an emulation environment for [this](https://github.com/rapid7/metasploit-framework/blob/master/modules/post/android/local/koffee.rb) Metasploit module. It composes of an emulated `micom` driver, an emulated `micomd` daemon and a userspace program that shows if an action is successfully triggered.

## Compile & Load

Make sure you have root access or at least can use `sudo`, then simply run `make all` under this directory. You might have to provide your password when the driver loads.
