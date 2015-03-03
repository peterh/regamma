package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/randr"
	"github.com/BurntSushi/xgb/screensaver"
	"github.com/BurntSushi/xgb/xproto"
)

var output = flag.String("output", "DisplayPort-0", "Reset the CRTC attached to this output")

type gamma struct {
	c     *xgb.Conn
	gamma *randr.GetCrtcGammaReply
	crtc  randr.Crtc
}

func (g *gamma) resetGamma() {
	randr.SetCrtcGamma(g.c, g.crtc, g.gamma.Size, g.gamma.Red, g.gamma.Green, g.gamma.Blue)
}

func screensaverListen(c *xgb.Conn, root xproto.Window) error {
	err := screensaver.Init(c)
	if err != nil {
		return err
	}
	ver, err := screensaver.QueryVersion(c, 1, 0).Reply()
	if err != nil {
		return err
	}
	if ver.ServerMajorVersion < 1 {
		return fmt.Errorf("Screensaver extesion is only version %d.%d (need at least 1.0)\n",
			ver.ServerMajorVersion, ver.ServerMinorVersion)
	}

	screensaver.SelectInput(c, xproto.Drawable(root), screensaver.EventNotifyMask)
	return nil
}

func findOutput(c *xgb.Conn, root xproto.Window) (randr.Crtc, error) {
again:
	list, err := randr.GetScreenResources(c, root).Reply()
	if err != nil {
		return 0, err
	}
	var getoutput []randr.GetOutputInfoCookie
	for _, outid := range list.Outputs {
		getoutput = append(getoutput, randr.GetOutputInfo(c, outid, list.ConfigTimestamp))
	}
	for _, cookie := range getoutput {
		r, err := cookie.Reply()
		if err != nil {
			return 0, err
		}
		if r.Status == randr.SetConfigInvalidConfigTime {
			goto again
		}
		if string(r.Name[:r.NameLen]) == *output {
			return r.Crtc, nil
		}
	}
	return 0, fmt.Errorf("no output named '%s'", *output)
}

func getGamma(c *xgb.Conn, root xproto.Window) (*gamma, error) {
	err := randr.Init(c)
	if err != nil {
		return nil, err
	}

	crtcid, err := findOutput(c, root)
	if err != nil {
		return nil, err
	}
	r, err := randr.GetCrtcGamma(c, crtcid).Reply()
	if err != nil {
		return nil, err
	}

	return &gamma{c: c, crtc: crtcid, gamma: r}, nil
}

func main() {
	flag.Parse()

	c, err := xgb.NewConn()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	root := xproto.Setup(c).Roots[c.DefaultScreen].Root
	g, err := getGamma(c, root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	err = screensaverListen(c, root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	for {
		xev, xerr := c.WaitForEvent()
		if xev != nil {
			switch xev := xev.(type) {
			case screensaver.NotifyEvent:
				if xev.Kind == screensaver.KindBlanked && xev.State == screensaver.StateOff {
					g.resetGamma()
				}
			default:
				fmt.Printf("Unexpected event: %#v\n", xev)
			}
		}
		if xerr != nil {
			fmt.Printf("Unexpected error: %#v\n", xerr)
		}
	}
}
