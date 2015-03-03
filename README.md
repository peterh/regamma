regamma
=======

regamma watches your screensaver state, and resets the gamma of the CRTC attached to a specified Output (the default is `DisplayPort-0`) whenever the X server wakes up the screen from DPMS sleep.

### Why?

regamma is intended to be used as a workaround for video cards and/or monitors that forget their gamma state during DPMS sleep. It was written because I'm using one such combination: A Dell P2210T attached to an AMD FirePro V3900 over DisplayPort using the stock drivers in Ubuntu 14.10.

### Usage

`go install` then place the following line in your `$HOME/.profile`:

```
nohup regamma > /dev/null &
```

### Caveats

If you change the gamma profile during your session, regamma will not notice and will restore your old gamma the next time your monitor wakes up from DPMS sleep. To avoid this problem, kill regamma before making adjustments, and re-start it after.

### License

regamma is licence under the X11 licence. See the file `COPYING`.
