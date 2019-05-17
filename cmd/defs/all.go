package defs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"git.parallelcoin.io/dev/9/cmd/ctl"
	"git.parallelcoin.io/dev/9/cmd/ll"
	"git.parallelcoin.io/dev/9/cmd/nine"
	"git.parallelcoin.io/dev/9/cmd/node"
	"git.parallelcoin.io/dev/9/cmd/walletmain"
	blockchain "git.parallelcoin.io/dev/9/pkg/chain"
	"git.parallelcoin.io/dev/9/pkg/chain/fork"
	"git.parallelcoin.io/dev/9/pkg/peer/connmgr"
	"git.parallelcoin.io/dev/9/pkg/util"
	"git.parallelcoin.io/dev/9/pkg/util/cl"
	"git.parallelcoin.io/dev/9/pkg/util/tcell"
	"git.parallelcoin.io/dev/9/pkg/util/tview"
	"github.com/btcsuite/go-socks/socks"
)

var DataDir string = filepath.Dir(util.AppDataDir("9", false))
var Networks = []string{"mainnet", "testnet", "simnet", "regtestnet"}
var NetParams = map[string]*nine.Params{
	"mainnet":    &nine.MainNetParams,
	"testnet":    &nine.TestNet3Params,
	"simnet":     &nine.SimNetParams,
	"regtestnet": &nine.RegressionNetParams,
}

// Log is the logger for node
var Log = cl.NewSubSystem("cmd/config", ll.DEFAULT)
var log = Log.Ch

var datadir *string = new(string)

const menutitle = "ⓟ parallelcoin 9 configuration CLI"

var iteminput *tview.InputField
var toggle *tview.Table

func NewIface() *Iface {
	return &Iface{Data: new(interface{})}
}

func (i *Iface) Get() interface{} {
	if i == nil {
		return nil
	}
	if i.Data == nil {
		return nil
	}
	return *i.Data
}

func (i *Iface) Put(in interface{}) *Iface {
	if i == nil {
		i = NewIface()
	}
	if i.Data == nil {
		i.Data = new(interface{})
	}
	*i.Data = in
	return i
}

func (r *Row) Bool() bool {
	return r.Value.Get().(bool)
}

func (r *Row) Int() int {
	return r.Value.Get().(int)
}

func (r *Row) Float() float64 {
	return r.Value.Get().(float64)
}

func (r *Row) Duration() time.Duration {
	return r.Value.Get().(time.Duration)
}

func (r *Row) Tag() string {
	return r.Value.Get().(string)
}

func (r *Row) Tags() []string {
	return r.Value.Get().([]string)
}

func (r *RowGenerators) RunAll(row *Row) {
	for _, x := range *r {
		x(row)
	}
}

func (r *CommandGenerators) RunAll() *Command {
	c := &Command{}
	for _, x := range *r {
		x(c)
	}
	return c
}

// GetSortedKeys returns the keys of a map in alphabetical order
func (r *CatJSON) GetSortedKeys() (out []string) {
	for i := range *r {
		out = append(out, i)
	}
	sort.Strings(out)
	return
}

func (r *CatsJSON) GetSortedKeys() (out []string) {
	for i := range *r {
		out = append(out, i)
	}
	sort.Strings(out)
	return
}

func (r *Cats) GetSortedKeys() (out []string) {
	for i := range *r {
		out = append(out, i)
	}
	sort.Strings(out)
	return
}

func (r Cat) GetSortedKeys() (out []string) {
	for i := range r {
		out = append(out, i)
	}
	sort.Strings(out)
	return
}

func (r *Tokens) GetSortedKeys() (out []string) {
	for i := range *r {
		out = append(out, i)
	}
	sort.Strings(out)
	return
}

func (r *Commands) GetSortedKeys() (out []string) {
	for i := range *r {
		out = append(out, i)
	}
	sort.Strings(out)
	return
}

// GenAddr returns a validator with a set default port assumed if one is not present
func GenAddr(name string, port int) func(r *Row, in interface{}) bool {
	return func(r *Row, in interface{}) bool {
		var s *string
		switch I := in.(type) {
		case string:
			s = &I
		case *string:
			s = I
		case nil:
			r.Value.Put(nil)
			return true
		default:
			return false
		}
		if s == nil {
			r.Value.Put(nil)
			return true
		}
		if *s == "" {
			s = nil
			r.Value.Put(nil)
			return true
		}
		h, p, err := net.SplitHostPort(*s)
		if err != nil {
			*s = net.JoinHostPort(*s, fmt.Sprint(port))
		} else {
			n, e := strconv.Atoi(p)
			if e == nil {
				if n < 1025 || n > 65535 {
					return false
				}
			} else {
				return false
				// p = ""
			}
			if p == "" {
				p = fmt.Sprint(port)
				*s = net.JoinHostPort(h, p)
			} else {
				*s = net.JoinHostPort(h, p)
			}
		}
		if r != nil {
			r.Value.Put(*s)
			r.String = *s
			r.App.SaveConfig()
		}
		return true
	}
}

// GenAddrs returns a validator with a set default port assumed if one is not present
func GenAddrs(name string, port int) func(r *Row, in interface{}) bool {
	return func(r *Row, in interface{}) bool {
		var s []string
		existing, ok := r.Value.Get().([]string)
		if !ok {
			existing = []string{}
		}
		switch I := in.(type) {
		case string:
			s = append(s, I)
		case *string:
			s = append(s, *I)
		case []string:
			s = I
		case *[]string:
			s = *I
		case []interface{}:
			for _, x := range I {
				so, ok := x.(string)
				if ok {
					s = append(s, so)
				}
			}
		case nil:
			return false
		default:
			fmt.Println(name, port, "invalid type", in, reflect.TypeOf(in))
			return false
		}
		for _, sse := range s {
			h, p, e := net.SplitHostPort(sse)
			if e != nil {
				sse = net.JoinHostPort(sse, fmt.Sprint(port))
			} else {
				n, e := strconv.Atoi(p)
				if e == nil {
					if n < 1025 || n > 65535 {
						fmt.Println(name, port, "port out of range")
						return false
					}
				} else {
					fmt.Println(name, port, "port not an integer")
					return false
				}
				if p == "" {
					p = fmt.Sprint(port)
				}
				sse = net.JoinHostPort(h, p)
			}
			existing = append(existing, sse)
		}
		if r != nil {
			// eliminate duplicates
			tmpMap := make(map[string]struct{})
			for _, x := range existing {
				tmpMap[x] = struct{}{}
			}
			existing = []string{}
			for i := range tmpMap {
				existing = append(existing, i)
			}
			sort.Strings(existing)
			r.Value.Put(existing)
			r.String = fmt.Sprint(existing)
			r.App.SaveConfig()
		}
		return true
	}
}

func getAlgoOptions() (options []string) {
	var modernd = "random"
	for _, x := range fork.P9AlgoVers {
		options = append(options, x)
	}
	options = append(options, modernd)
	sort.Strings(options)
	return
}

// Valid is a collection of validator functions for the different types used
// in a configuration. These functions optionally can accept a *Row and with
// this they assign the validated, parsed value into the Value slot.
var Valid = struct {
	File, Dir, Port, Bool, Int, Tag, Tags, Algo, Float, Duration, Net,
	Level func(*Row, interface{}) bool
}{
	File: func(r *Row, in interface{}) bool {
		var s *string
		switch I := in.(type) {
		case string:
			s = &I
		case *string:
			s = I
		default:
			return false
		}
		if len(*s) > 0 {
			ss := CleanAndExpandPath(*s, *datadir)
			if r != nil {
				r.String = fmt.Sprint(ss)
				if r.Value == nil {
					r.Value = NewIface()
				}
				r.Value.Put(ss)
				r.App.SaveConfig()
				return true
			} else {
				return false
			}
		}
		return false
	},
	Dir: func(r *Row, in interface{}) bool {
		var s *string
		switch I := in.(type) {
		case string:
			s = &I
		case *string:
			s = I
		default:
			return false
		}
		if len(*s) > 0 {
			ss := CleanAndExpandPath(*s, *datadir)
			if r != nil {
				r.String = fmt.Sprint(ss)
				if r.Value == nil {
					r.Value = NewIface()
				}
				r.Value.Put(ss)
				r.App.SaveConfig()
				return true
			} else {
				return false
			}
		}
		return false
	},
	Port: func(r *Row, in interface{}) bool {
		var s string
		var ii int
		isString := false
		switch I := in.(type) {
		case string:
			s = I
			isString = true
		case *string:
			s = *I
			isString = true
		case int:
			ii = I
		case *int:
			ii = *I
		default:
			return false
		}
		if isString {
			n, e := strconv.Atoi(s)
			if e != nil {
				return false
			}
			ii = n
		}
		if ii < 1025 || ii > 65535 {
			return false
		}
		if r != nil {
			r.Value.Put(ii)
			r.String = fmt.Sprint(ii)
			r.App.SaveConfig()
		}
		return true
	},
	Bool: func(r *Row, in interface{}) bool {
		var sb string
		var b bool
		switch I := in.(type) {
		case string:
			sb = I
			if strings.ToUpper(sb) == "TRUE" {
				b = true
				goto boolout
			}
			if strings.ToUpper(sb) == "FALSE" {
				b = false
				goto boolout
			}
		case *string:
			sb = *I
			if strings.ToUpper(sb) == "TRUE" {
				b = true
				goto boolout
			}
			if strings.ToUpper(sb) == "FALSE" {
				b = false
				goto boolout
			}
		case bool:
			b = I
		case *bool:
			b = *I
		default:
			return false
		}
	boolout:
		if r != nil {
			r.String = fmt.Sprint(b)
			r.Value.Put(b)
			r.App.SaveConfig()
		}
		return true
	},
	Int: func(r *Row, in interface{}) bool {
		var s string
		var ii int
		isString := false
		switch I := in.(type) {
		case string:
			s = I
			isString = true
		case *string:
			s = *I
			isString = true
		case int:
			ii = I
		case *int:
			ii = *I
		default:
			return false
		}
		if isString {
			n, e := strconv.Atoi(s)
			if e != nil {
				return false
			}
			ii = n
		}
		if r != nil {
			r.String = fmt.Sprint(ii)
			//r.Value =
			r.Value.Put(ii)
			r.App.SaveConfig()
		}
		return true
	},
	Tag: func(r *Row, in interface{}) bool {
		var s string
		switch I := in.(type) {
		case string:
			s = I
		case *string:
			s = *I
		default:
			return false
		}
		s = strings.TrimSpace(s)
		if len(s) < 1 {
			return false
		}
		if r != nil {
			r.Value.Put(s)
			r.String = fmt.Sprint(s)
			r.App.SaveConfig()
		}
		return true
	},
	Tags: func(r *Row, in interface{}) bool {
		var s []string
		existing, ok := r.Value.Get().([]string)
		if !ok {
			existing = []string{}
		}
		switch I := in.(type) {
		case string:
			s = append(s, I)
		case *string:
			s = append(s, *I)
		case []string:
			s = I
		case *[]string:
			s = *I
		case []interface{}:
			for _, x := range I {
				so, ok := x.(string)
				if ok {
					s = append(s, so)
				}
			}
		case nil:
			return false
		default:
			fmt.Println("invalid type", in, reflect.TypeOf(in))
			return false
		}
		for _, sse := range s {
			existing = append(existing, sse)
		}
		if r != nil {
			// eliminate duplicates
			tmpMap := make(map[string]struct{})
			for _, x := range existing {
				tmpMap[x] = struct{}{}
			}
			existing = []string{}
			for i := range tmpMap {
				existing = append(existing, i)
			}
			sort.Strings(existing)
			r.Value.Put(existing)
			r.String = fmt.Sprint(existing)
			r.App.SaveConfig()
		}
		return true
	},
	Algo: func(r *Row, in interface{}) bool {
		var s string
		switch I := in.(type) {
		case string:
			s = I
		case *string:
			s = *I
		default:
			return false
		}
		var o string
		options := getAlgoOptions()
		for _, x := range options {
			if s == x {
				o = s
			}
		}
		if o == "" {
			rnd := "random"
			o = rnd
		}
		if r != nil {
			r.String = fmt.Sprint(o)
			r.Value.Put(o)
			r.App.SaveConfig()
		}
		return true
	},
	Float: func(r *Row, in interface{}) bool {
		var s string
		var f float64
		isString := false
		switch I := in.(type) {
		case string:
			s = I
			isString = true
		case *string:
			s = *I
			isString = true
		case float64:
			f = I
		case *float64:
			f = *I
		default:
			return false
		}
		if isString {
			ff, e := strconv.ParseFloat(s, 64)
			if e != nil {
				return false
			}
			f = ff
		}
		if r != nil {
			r.Value.Put(f)
			r.String = fmt.Sprint(f)
			r.App.SaveConfig()
		}
		return true
	},
	Duration: func(r *Row, in interface{}) bool {
		var s string
		var t time.Duration
		isString := false
		switch I := in.(type) {
		case string:
			s = I
			isString = true
		case *string:
			s = *I
			isString = true
		case time.Duration:
			t = I
		case *time.Duration:
			t = *I
		default:
			return false
		}
		if isString {
			dd, e := time.ParseDuration(s)
			if e != nil {
				return false
			}
			t = dd
		}
		if r != nil {
			r.String = fmt.Sprint(t)
			r.Value.Put(t)
			r.App.SaveConfig()
		}
		return true
	},
	Net: func(r *Row, in interface{}) bool {
		var sn string
		switch I := in.(type) {
		case string:
			sn = I
		case *string:
			sn = *I
		default:
			return false
		}
		found := false
		for _, x := range Networks {
			if x == sn {
				found = true
				*nine.ActiveNetParams = *NetParams[x]
			}
		}
		if r != nil && found {
			r.String = fmt.Sprint(sn)
			r.Value.Put(sn)
			r.App.SaveConfig()
		}
		return found
	},
	Level: func(r *Row, in interface{}) bool {
		var sl string
		switch I := in.(type) {
		case string:
			sl = I
		case *string:
			sl = *I
		default:
			return false
		}
		found := false
		for x := range cl.Levels {
			if x == sl {
				found = true
			}
		}
		if r != nil && found {
			r.String = fmt.Sprint(sl)
			r.Value.Put(sl)
			r.App.SaveConfig()
		}
		return found
	},
}

func MainColor() tcell.Color {
	return tcell.NewRGBColor(64, 64, 64)
}

func DimColor() tcell.Color {
	return tcell.NewRGBColor(48, 48, 48)
}

func PrelightColor() tcell.Color {
	return tcell.NewRGBColor(32, 32, 32)
}

func TextColor() tcell.Color {
	return tcell.NewRGBColor(216, 216, 216)
}

func BackgroundColor() tcell.Color {
	return tcell.NewRGBColor(16, 16, 16)
}

func Run(args []string, tokens Tokens, app *App) int {
	var cattable *tview.Table
	var cattablewidth int

	var activepage *tview.Flex
	var inputhandler func(event *tcell.EventKey) *tcell.EventKey
	var cat, itemname string

	// tapp pulls everything together to create the configuration interface
	tapp := tview.NewApplication()

	// titlebar tells the user what app they are using
	titlebar := tview.NewTextView().
		SetTextColor(TextColor()).
		SetText(menutitle)
	titlebar.Box.SetBackgroundColor(MainColor())

	coverbox := tview.NewTextView()
	coverbox.
		SetTextColor(TextColor())
	coverbox.Box.
		SetBorder(false).
		SetBackgroundColor(BackgroundColor())
	coverbox.SetBorderPadding(1, 1, 2, 2)
	// coverbox.SetBorder(true)

	roottable, roottablewidth := genMenu("launch", "configure", "reinitialize")
	activateTable(roottable)

	launchmenutexts := []string{"node", "wallet", "shell"}
	launchtable, launchtablewidth := genMenu(launchmenutexts...)
	prelightTable(launchtable)

	catstable, catstablewidth := genMenu(app.Cats.GetSortedKeys()...)
	prelightTable(catstable)

	menuflex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(roottable, roottablewidth, 1, true).
		AddItem(coverbox, 0, 1, false)
	menuflex.Box.SetBackgroundColor(BackgroundColor())

	var leftExitActive bool
	var confirm *tview.Flex
	roottable.SetSelectionChangedFunc(func(y, x int) {
		leftExitActive = false
		coverbox.SetText(
			"",
		)
		menuflex.
			RemoveItem(coverbox).
			RemoveItem(launchtable).
			RemoveItem(catstable).
			RemoveItem(cattable).
			RemoveItem(confirm)
		switch y {
		case 0, 3:
			menuflex.
				AddItem(coverbox, 0, 1, true)
		case 1:
			menuflex.
				AddItem(launchtable, launchtablewidth, 1, true).
				AddItem(coverbox, 0, 1, true)
		case 2:
			menuflex.
				AddItem(catstable, catstablewidth, 1, true)
			if cattable != nil {
				lastTable(cattable)
				menuflex.AddItem(cattable, cattablewidth, 1, true)
			}
			menuflex.AddItem(coverbox, 0, 1, true)
		}
	})
	var resetbutton int
	var toggleResetButton = func() int {
		if resetbutton == 0 {
			resetbutton = 1
		} else {
			resetbutton = 0
		}
		return resetbutton
	}
	var factoryResetFunc = func() {
		confirm = tview.NewFlex()
		confirm.SetDirection(tview.FlexRow)
		confirm.SetBorderPadding(1, 1, 2, 2)
		resettext := tview.NewTextView()
		resettext.SetText("all custom configurations will be lost, are you sure?")
		resettext.SetBorderPadding(1, 1, 2, 2)
		resettext.SetWordWrap(true)
		resettext.SetTextAlign(tview.AlignCenter)
		resettext.Box.SetBackgroundColor(MainColor())
		resetform := tview.NewForm()
		resetform.Box.SetBackgroundColor(MainColor())
		resetform.SetButtonsAlign(tview.AlignCenter)
		resetform.SetButtonBackgroundColor(MainColor())
		resetform.SetButtonTextColor(TextColor())
		eventcap := func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyTab:
				tapp.SetFocus(resetform.GetButton(toggleResetButton()))
			case tcell.KeyRight, tcell.KeyLeft:
				tapp.SetFocus(resetform.GetButton(toggleResetButton()))
			case tcell.KeyEsc:
				resetform.Blur()
				roottable.Select(3, 0)
				tapp.SetFocus(roottable)
				menuflex.RemoveItem(confirm)
				menuflex.AddItem(coverbox, 0, 1, false)
				return &tcell.EventKey{}
			}
			return event
		}
		resetform.AddButton("cancel", func() {
			menuflex.RemoveItem(confirm)
			tapp.SetFocus(roottable)
		})
		resetform.AddButton("reset to factory settings", func() {
			for _, x := range app.Cats {
				for _, z := range x {
					z.Init(z)
				}
			}
			resettext.SetText("CONFIRMED\n\nfactory reset completed")
			confirm.RemoveItem(resetform)
			// resetform.RemoveButton(1)
			tapp.ForceDraw()
			time.Sleep(time.Second)
			menuflex.RemoveItem(confirm)
			tapp.SetFocus(roottable)
		})
		resetform.SetInputCapture(eventcap)
		resetform.GetButton(0).SetInputCapture(eventcap)
		resetform.GetButton(1).SetInputCapture(eventcap)
		confirm.AddItem(resettext, 5, 0, false)
		confirm.AddItem(resetform, 3, 0, true)
		menuflex.AddItem(confirm, 0, 1, true)
		tapp.SetFocus(confirm)
	}
	roottable.SetSelectedFunc(func(y, x int) {
		menuflex.RemoveItem(coverbox)
		if cattable != nil {
			menuflex.RemoveItem(cattable)
		}
		switch y {
		case 0:
			tapp.Stop()
		case 1:
			activatedTable(roottable)
			activateTable(launchtable)
			menuflex.AddItem(coverbox, 0, 1, true)
			tapp.SetFocus(launchtable)
		case 2:
			activatedTable(roottable)
			activateTable(catstable)
			if cattable != nil {
				menuflex.AddItem(cattable, cattablewidth, 0, false)
				prelightTable(cattable)
			}
			menuflex.AddItem(coverbox, 0, 1, true)
			tapp.SetFocus(catstable)
		case 3:
			factoryResetFunc()
		}
	})
	roottable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		menuflex.RemoveItem(coverbox)
		roottable.GetCell(0, 0).SetText("<")
		menuflex.
			RemoveItem(cattable).
			RemoveItem(coverbox)
		switch event.Key() {
		case tcell.KeyRight, tcell.KeyTab:
			leftExitActive = false
			y, _ := roottable.GetSelection()
			switch y {
			case 1:
				activatedTable(roottable)
				activateTable(launchtable)
				menuflex.AddItem(coverbox, 0, 1, true)
				tapp.SetFocus(launchtable)
			case 2:
				activatedTable(roottable)
				activateTable(catstable)
				if cattable != nil {
					menuflex.AddItem(cattable, cattablewidth, 0, false)
					prelightTable(cattable)
				}
				menuflex.AddItem(coverbox, 0, 1, true)
				tapp.SetFocus(catstable)
			case 3:
				factoryResetFunc()
			}
		case tcell.KeyLeft, tcell.KeyEsc:
			y, _ := roottable.GetSelection()
			if y == 0 {
				if !leftExitActive {
					roottable.GetCell(0, 0).SetText("< exit")
					leftExitActive = true
				} else {
					tapp.Stop()
				}
			} else {
				roottable.Select(0, 0)
			}
		}
		return event
	})

	launchtable.SetSelectionChangedFunc(func(y, x int) {
		switch y {
		case 0:
			menuflex.
				RemoveItem(coverbox).
				RemoveItem(cattable).
				RemoveItem(catstable).
				AddItem(coverbox, 0, 1, false)
			coverbox.SetText("")
		case 1:
			coverbox.SetText("run a full peer to peer parallelcoin node")
		case 2:
			coverbox.SetText("\nrun a wallet server (requires a full node)")
		case 3:
			coverbox.SetText("\n\nrun a combined wallet/full node")
		}
	})
	launchtable.SetSelectedFunc(func(y, x int) {
		switch y {
		case 0:
			prelightTable(launchtable)
			activateTable(roottable)
			tapp.SetFocus(roottable)
			return
		case 1:
			tapp.Stop()
			fmt.Println("starting up", launchmenutexts[y-1])
			app.Commands[launchmenutexts[y-1]].Handler(args, tokens, app)
		case 2:
			tapp.Stop()
			fmt.Println("starting up", launchmenutexts[y-1])
			app.Commands[launchmenutexts[y-1]].Handler(args, tokens, app)
		case 3:
			tapp.Stop()
			fmt.Println("starting up", launchmenutexts[y-1])
			app.Commands[launchmenutexts[y-1]].Handler(args, tokens, app)
		}
	})
	launchtable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyLeft, tcell.KeyEsc:
			prelightTable(launchtable)
			activateTable(roottable)
			tapp.SetFocus(roottable)
		}
		return event
	})

	saveConfig := func() {
		ddir, ok := app.Cats["app"]["datadir"].Get().(string)
		if ok {
			configFile := CleanAndExpandPath(filepath.Join(
				ddir, "config"), "")
			if EnsureDir(configFile) {
			}
			fh, err := os.Create(configFile)
			if err != nil {
				panic(err)
			}
			j, e := json.MarshalIndent(app, "", "\t")
			if e != nil {
				panic(e)
			}
			_, err = fmt.Fprint(fh, string(j))
			if err != nil {
				panic(err)
			}
		}
	}

	var genPage func(cat, item string, active bool, app *App,
		editoreventhandler func(event *tcell.EventKey) *tcell.EventKey, idx int) (out *tview.Flex)

	inputhandler = func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			menuflex.
				RemoveItem(coverbox).
				RemoveItem(activepage)
			activepage = genPage(cat, itemname, false, app, inputhandler, 0)
			menuflex.AddItem(activepage, 0, 1, true)
			prelightTable(roottable)
			activatedTable(catstable)
			activateTable(cattable)
			tapp.SetFocus(cattable)
		default:
		}
		return event
	}

	genPage = func(cat, item string, active bool, app *App,
		editoreventhandler func(event *tcell.EventKey) *tcell.EventKey, idx int) (out *tview.Flex) {
		currow := app.Cats[cat][item]
		var darkness, lightness tcell.Color
		if active {
			darkness = MainColor()
			lightness = TextColor()
		} else {
			darkness = PrelightColor()
			lightness = MainColor()
		}
		out = tview.NewFlex().SetDirection(tview.FlexRow)
		heading := tview.NewTextView().
			SetText(fmt.Sprintf("%s.%s", cat, item))
		heading.
			SetTextColor(lightness).
			SetBackgroundColor(darkness).
			SetBorderPadding(0, 0, 1, 1)
		out.
			SetBorderPadding(1, 1, 1, 1).
			SetBackgroundColor(darkness)
		out.AddItem(heading, 2, 0, false)
		infoblock := tview.NewTextView()
		infoblock.
			SetWordWrap(true).
			SetTextColor(lightness).
			SetBorderPadding(1, 0, 1, 1).
			SetBackgroundColor(darkness)
		def := currow.Default
		defstring := ""
		if def != nil {
			defstring = fmt.Sprintf("default value: %v", def.Get())
		} else {
			defstring = "" //"this value has no default"
		}
		infostring := fmt.Sprintf(
			"%v\n\n%s",
			currow.Usage, defstring,
		)
		if min, ok := currow.Min.Get().(int); ok {
			infostring += fmt.Sprint("\nminimum value: ", min)
		}
		if max, ok := currow.Max.Get().(int); ok {
			infostring += fmt.Sprint("\nmaximum value: ", max)
		}
		itemtype := currow.Type
		infostring =
			"<esc>     to cancel\n\n" + infostring
		switch currow.Type {
		case "int", "float", "duration":
			infostring =
				"<ctrl-z>  to reset to default\n" +
					infostring
		case "string", "port":
			infostring =
				"<ctrl-u>  to clear\n" +
					"<ctrl-z>  to reset to default\n" +
					infostring
		default:
		}
		infoblock.SetText(infostring)
		switch itemtype {
		case "string", "int", "float", "duration", "port":

			iteminput = tview.NewInputField()
			iteminput.
				SetFieldTextColor(darkness).
				SetFieldBackgroundColor(lightness).
				SetBackgroundColor(lightness).
				SetBorderPadding(1, 1, 1, 1)
			val := currow.Value
			if val != nil {
				vv := val.Get()
				outstring := ""
				if vv != nil {
					switch ov := vv.(type) {
					case int:
						outstring = fmt.Sprintf("%8d", ov)
					case float64:
						switch itemtype {
						case "float":
							os := fmt.Sprintf("%0f", ov)
							os = strings.TrimRight(os, "0")
							if strings.HasSuffix(os, ".") {
								os += "0"
							}
							outstring = os
						case "int", "port":
							outint := int(ov)
							outstring = fmt.Sprintf("%8d", outint)
						case "duration":
							outdur := time.Duration(int(ov))
							outstring = fmt.Sprintf("%v", outdur)
						}
					case time.Duration:
						outstring = fmt.Sprintf("%v", ov)
					default:
						outstring = fmt.Sprint(ov)
					}
					iteminput.SetText(strings.TrimSpace(outstring))
				}
			}
			var canceller func(rw *Row) func(event *tcell.EventKey) *tcell.EventKey
			canceller = func(rw *Row) func(event *tcell.EventKey) *tcell.EventKey {
				return func(event *tcell.EventKey) *tcell.EventKey {
					switch {
					case event.Key() == tcell.KeyCtrlU:
						switch itemtype {
						case "int":
							rw.Value.Put(rw.Default.Get())
						case "float":
							rw.Value.Put(rw.Default.Get())
						case "duration":
							rw.Value.Put(rw.Default.Get())
						default:
							rw.Value.Put(nil)
						}
					case event.Key() == tcell.KeyCtrlZ:
						rw.Value.Put(rw.Default.Get())
					default:
						return editoreventhandler(event)
					}
					menuflex.
						RemoveItem(coverbox).
						RemoveItem(activepage)
					itemname = item
					activepage = genPage(cat, itemname, false, app, canceller(rw), 0)
					menuflex.AddItem(activepage, 0, 1, true)
					prelightTable(roottable)
					activatedTable(catstable)
					activateTable(cattable)
					tapp.SetFocus(cattable)
					saveConfig()
					return event
				}
			}
			iteminput.SetInputCapture(canceller(currow))
			snackbar := tview.NewTextView()
			iteminput.SetDoneFunc(func(key tcell.Key) {
				rrr := currow
				rw := rrr
				if key == tcell.KeyEnter || key == tcell.KeyTab {
					s := iteminput.GetText()
					if s == "" {
						switch itemtype {
						case "int":
							rw.Value.Put(0)
						case "float":
							rw.Value.Put(0.0)
						case "duration":
							rw.Value.Put(0 * time.Second)
						default:
							rw.Value.Put(nil)
						}
						saveConfig()
					} else {
						isvalid := rw.Validate(rw, &s)
						if !isvalid {
							snackbar.SetBackgroundColor(tcell.ColorOrange)
							snackbar.SetTextColor(tcell.ColorRed)
							snackbar.SetText("input is not valid for this field")
							out.RemoveItem(infoblock).RemoveItem(snackbar)
							out.AddItem(snackbar, 1, 1, false)
							out.AddItem(infoblock, 0, 1, false)
							return
						} else {
							// rw.Validate(rw, s)
							// rw.Value.Put(s)
							saveConfig()
							out.RemoveItem(snackbar)
						}
					}
					menuflex.
						RemoveItem(coverbox).
						RemoveItem(activepage)
					itemname = item
					activepage = genPage(cat, itemname, false, app, inputhandler, 0)
					menuflex.AddItem(activepage, 0, 1, true)
					prelightTable(roottable)
					activatedTable(catstable)
					activateTable(cattable)
					tapp.SetFocus(cattable)
				}
			})
			out.AddItem(iteminput, 3, 0, true)
		case "bool":
			rw := currow
			toggle = tview.NewTable()
			toggle.SetBorderPadding(1, 1, 2, 2)
			toggle.SetBackgroundColor(lightness)
			def := currow.Default.Get().(bool)
			if def {
				toggle.
					SetCell(0, 0, tview.NewTableCell("false").SetTextColor(darkness)).
					SetCell(1, 0, tview.NewTableCell("true (default)").SetTextColor(darkness))
			} else {
				toggle.
					SetCell(0, 0, tview.NewTableCell("false (default)").SetTextColor(darkness)).
					SetCell(1, 0, tview.NewTableCell("true").SetTextColor(darkness))
			}
			curropt := 0
			curr := currow
			if curr.Bool() {
				curropt = 1
			}
			toggle.
				SetSelectable(true, true).
				Select(curropt, 0)
			toggle.SetBackgroundColor(lightness)
			toggle.SetInputCapture(editoreventhandler)
			toggle.SetSelectedFunc(func(y, x int) {
				menuflex.
					RemoveItem(coverbox).
					RemoveItem(activepage)
				switch y {
				case 0:
					rw.Put(false)
					itemname = item
					activepage = genPage(cat, itemname, false, app, inputhandler, y)
					menuflex.AddItem(activepage, 0, 1, true)
					prelightTable(roottable)
					activatedTable(catstable)
					activateTable(cattable)
					tapp.SetFocus(cattable)
				case 1:
					rw.Put(true)
					itemname = item
					activepage = genPage(cat, itemname, false, app, inputhandler, y)
					menuflex.AddItem(activepage, 0, 1, true)
					prelightTable(roottable)
					activatedTable(catstable)
					activateTable(cattable)
					tapp.SetFocus(cattable)

				default:
				}
				saveConfig()
			})
			out.AddItem(toggle, 4, 0, true)
		case "options":
			rw := currow
			var toggle = tview.NewTable()
			toggle.SetBorderPadding(1, 1, 1, 1)
			def := currow.Default.Get().(string)
			curr := currow.Value.Get().(string)
			curropt := 0
			sort.Strings(currow.Opts)
			for i, x := range currow.Opts {
				itemtext := x
				if x == def {
					itemtext += " (default)"
				}
				if x == curr {
					curropt = i
				}
				toggle.
					SetCell(i, 0, tview.NewTableCell(itemtext).
						SetTextColor(darkness).SetBackgroundColor(lightness))
			}
			toggle.
				SetSelectable(true, true).
				Select(curropt, 0)
			toggle.SetBackgroundColor(lightness)
			toggle.SetInputCapture(editoreventhandler)
			toggle.SetSelectedFunc(func(y, x int) {
				menuflex.
					RemoveItem(coverbox).
					RemoveItem(activepage)
				rw.Put(currow.Opts[y])
				saveConfig()
				itemname = item
				activepage = genPage(cat, itemname, false, app, inputhandler, y)
				menuflex.AddItem(activepage, 0, 1, true)
				prelightTable(roottable)
				activatedTable(catstable)
				activateTable(cattable)
				tapp.SetFocus(cattable)
			})
			out.AddItem(toggle, len(currow.Opts)+2, 0, true)
		case "stringslice":
			var slice = tview.NewTable()
			slice.SetBorderPadding(1, 1, 1, 1)
			var def string
			defIface := currow.Default.Get()
			switch defIface.(type) {
			case string:
				def = currow.Default.Get().(string)
			case nil:
			default:
			}
			var curr string
			currIface := currow.Value.Get()
			switch currIface.(type) {
			case string:
				curr = currIface.(string)
			case nil:
			default:
			}
			curropt := 0
			slicevalue, ok := currow.Get().([]string)
			if ok {
				for i, x := range slicevalue {
					itemtext := x
					if x == def {
						itemtext += " (default)"
					}
					if x == curr {
						curropt = i
					}
					slice.
						SetCell(i, 0, tview.NewTableCell("⌦").
							SetTextColor(darkness).SetBackgroundColor(lightness))
					slice.
						SetCell(i, 1, tview.NewTableCell(itemtext).
							SetTextColor(darkness).SetBackgroundColor(lightness))
				}
			}
			slice.
				SetCell(len(slicevalue), 1, tview.NewTableCell("add new").
					SetTextColor(darkness).SetBackgroundColor(lightness))
			slice.
				SetCell(len(slicevalue), 0, tview.NewTableCell("").
					SetTextColor(darkness).SetBackgroundColor(lightness).
					SetSelectable(false))
			slice.
				SetCell(len(slicevalue)+1, 1, tview.NewTableCell("set defaults").
					SetTextColor(darkness).SetBackgroundColor(lightness))
			slice.
				SetCell(len(slicevalue)+1, 0, tview.NewTableCell("").
					SetTextColor(darkness).SetBackgroundColor(lightness).
					SetSelectable(false))
			slice.
				SetCell(len(slicevalue)+2, 1, tview.NewTableCell("back").
					SetTextColor(darkness).SetBackgroundColor(lightness))
			slice.
				SetCell(len(slicevalue)+2, 0, tview.NewTableCell("").
					SetTextColor(darkness).SetBackgroundColor(lightness).
					SetSelectable(false))
			input := tview.NewInputField()
			snackbar := tview.NewTextView()
			inputDoneGen := func(idx int) func(key tcell.Key) {
				return func(key tcell.Key) {
					rrr := currow
					rw := rrr
					rwv, ok := rw.Value.Get().([]string)
					if !ok { // rwv = []string{}
					}
					if key == tcell.KeyEnter || key == tcell.KeyTab {
						s := input.GetText()
						if len(s) < 1 {
							// rw.Value.Put(nil)
						} else {
							if rw.Validate(rw, s) {
								// if idx >= len(rwv) {
								// 	rwv = append(rwv, s)
								// } else {
								// 	rwv[idx] = s
								// }
								// rw.Value.Put(rwv)
							} else {
								snackbar.SetBackgroundColor(tcell.ColorOrange)
								snackbar.SetTextColor(tcell.ColorRed)
								snackbar.SetText("input is not valid for this field")
								out.RemoveItem(infoblock).RemoveItem(snackbar)
								out.AddItem(snackbar, 1, 1, false)
								out.AddItem(infoblock, 0, 1, false)
								return
							}
							saveConfig()
							out.RemoveItem(snackbar)
						}

						// itemname = item
						// inputhandler = func(event *tcell.EventKey) *tcell.EventKey {
						// 	switch event.Key() {
						// 	case 13:
						// 		// pressed enter
						// 	case 27:
						// 		// pressed escape
						// 		menuflex.
						// 			RemoveItem(coverbox).
						// 			RemoveItem(activepage)
						// 		// itemname = item
						// 		activepage = genPage(cat, itemname, false, app, inputhandler, idx)
						// 		menuflex.AddItem(activepage, 0, 1, true)
						// 		prelightTable(roottable)
						// 		activatedTable(catstable)
						// 		activateTable(cattable)
						// 		tapp.SetFocus(cattable)
						// 	}
						// 	return event
						// }

						menuflex.
							RemoveItem(coverbox).
							RemoveItem(activepage)
						itemname = item
						activepage = genPage(cat, itemname, true, app, inputhandler, idx)
						menuflex.AddItem(activepage, 0, 1, true)
						lastTable(roottable)
						prelightTable(catstable)
						activatedTable(cattable)
						slice.Select(idx, 1)
						tapp.SetFocus(activepage)
					}
					if key == tcell.KeyEsc {
						menuflex.
							RemoveItem(coverbox).
							RemoveItem(activepage)
						itemname = item
						activepage = genPage(cat, itemname, true, app, inputhandler, len(rwv))
						menuflex.AddItem(activepage, 0, 1, true)
						lastTable(roottable)
						prelightTable(catstable)
						activatedTable(cattable)
						tapp.SetFocus(activepage)
						// return event //&tcell.EventKey{}
					}

				}
			}
			slice.SetSelectedFunc(func(y, x int) {
				switch {
				// create new
				case y == len(slicevalue):
					// pop up the new item editor
					out.RemoveItem(infoblock)
					input.SetBackgroundColor(lightness)
					input.SetLabel("new> ")
					input.SetLabelColor(darkness)
					input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
						if event.Key() == 27 {
							menuflex.
								RemoveItem(coverbox).
								RemoveItem(activepage)
							itemname = item
							activepage = genPage(cat, itemname, true, app, inputhandler, y)
							menuflex.AddItem(activepage, 0, 1, true)
							lastTable(roottable)
							prelightTable(catstable)
							activatedTable(cattable)
							tapp.SetFocus(activepage)
							// return event // &tcell.EventKey{}
						}
						return event
					})
					input.SetDoneFunc(inputDoneGen(y))
					out.AddItem(input, 1, 0, true).
						AddItem(infoblock, 0, 1, false)
					tapp.SetFocus(input)

				// set default
				case y == len(slicevalue)+1:
					currow.Init(currow)

					menuflex.
						RemoveItem(coverbox).
						RemoveItem(activepage)
					itemname = item
					activepage = genPage(cat, itemname, false, app, inputhandler, y)
					menuflex.AddItem(activepage, 0, 1, true)
					prelightTable(roottable)
					activatedTable(catstable)
					activateTable(cattable)
					tapp.SetFocus(cattable)

				// back
				case y == len(slicevalue)+2:
					menuflex.
						RemoveItem(coverbox).
						RemoveItem(activepage)
					itemname = item
					activepage = genPage(cat, itemname, false, app, inputhandler, y)
					menuflex.AddItem(activepage, 0, 1, true)
					prelightTable(roottable)
					activatedTable(catstable)
					activateTable(cattable)
					tapp.SetFocus(cattable)

					//existing
				default:
					rw := currow
					rwv, ok := rw.Value.Get().([]string)
					// column 0 is delete column 1 is edit
					// TODO: consolidate editor code from above with this
					if x == 0 {
						if ok {
							// deleted := rwv[y]
							rwv = append(rwv[:y], rwv[y+1:]...)
							rw.Value.Put(rwv)
							saveConfig()
							menuflex.
								RemoveItem(coverbox).
								RemoveItem(activepage)
							itemname = item
							activepage = genPage(cat, itemname, true, app, inputhandler, y)
							menuflex.AddItem(activepage, 0, 1, true)
							lastTable(roottable)
							prelightTable(catstable)
							activatedTable(cattable)
							tapp.SetFocus(activepage)
						} else {
							// rw.Value.Put([]string{})
						}
					} else {
						// pop up the item editor
						out.RemoveItem(infoblock)
						input.SetBackgroundColor(lightness)
						input.SetLabel("edit> ")
						input.SetLabelColor(darkness)
						if len(rwv) >= y {
							input.SetText(rwv[y])
						}
						input.SetDoneFunc(inputDoneGen(y))
						out.AddItem(input, 1, 0, true).
							AddItem(infoblock, 0, 1, false)
						tapp.SetFocus(input)

					}
				}
			})
			slice.
				SetSelectable(true, true).
				Select(curropt, 1)
			slice.SetBackgroundColor(lightness)
			slice.SetInputCapture(editoreventhandler)
			slice.Select(len(slicevalue), 1)
			out.AddItem(slice, len(slicevalue)+5, 0, true)
		}
		out.AddItem(infoblock, 0, 1, false)

		return
	}
	catstable.SetSelectionChangedFunc(func(y, x int) {
		itemname = ""
		menuflex.
			RemoveItem(activepage).
			RemoveItem(coverbox).
			RemoveItem(cattable)
		if y == 0 {
			cat = strings.TrimSpace(catstable.GetCell(y, x).Text)
			menuflex.
				AddItem(coverbox, 0, 1, true)
			return
		}
		cat = app.Cats.GetSortedKeys()[y-1]
		ckeys := app.Cats[cat].GetSortedKeys()
		var catkeys []string
		for _, x := range ckeys {
			if !(cat == "app" && x == "datadir") {
				catkeys = append(catkeys, x)
			}
		}
		cattable, cattablewidth = genMenu(catkeys...)
		prelightTable(cattable)
		cattable.SetSelectedFunc(func(y, x int) {
			menuflex.
				RemoveItem(activepage).
				RemoveItem(coverbox)
			if y == 0 {
				activatedTable(roottable)
				prelightTable(cattable)
				activateTable(catstable)
				menuflex.
					AddItem(coverbox, 0, 1, true)
				tapp.SetFocus(catstable)
			} else {
				lastTable(roottable)
				prelightTable(catstable)
				activatedTable(cattable)
				var catkeys []string
				for _, x := range app.Cats[cat].GetSortedKeys() {
					if !(cat == "app" && x == "datadir") {
						catkeys = append(catkeys, x)
					}
				}
				itemname = catkeys[y-1]
				activepage = genPage(cat, itemname, true, app, inputhandler, 0)
				menuflex.AddItem(activepage, 0, 1, true)

				tapp.SetFocus(activepage)
			}
		})
		cattable.SetSelectionChangedFunc(func(y, x int) {
			menuflex.
				RemoveItem(coverbox).
				RemoveItem(activepage)
			if y == 0 {
				menuflex.AddItem(coverbox, 0, 1, false)
			} else {
				var catkeys []string
				for _, x := range app.Cats[cat].GetSortedKeys() {
					if !(cat == "app" && x == "datadir") {
						catkeys = append(catkeys, x)
					}
				}
				itemname = catkeys[y-1]
				activepage = genPage(cat, itemname, false, app, nil, y)
				menuflex.AddItem(activepage, 0, 1, true)
			}
		})
		cattable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyRight, tcell.KeyTab:
				menuflex.
					RemoveItem(activepage).
					RemoveItem(coverbox)
				y, _ := cattable.GetSelection()
				if y == 0 {
					break
				}
				lastTable(roottable)
				prelightTable(catstable)
				activatedTable(cattable)
				var catkeys []string
				for _, x := range app.Cats[cat].GetSortedKeys() {
					if !(cat == "app" && x == "datadir") {
						catkeys = append(catkeys, x)
					}
				}
				itemname = catkeys[y-1]
				activepage = genPage(cat, itemname, true, app, inputhandler, 0)
				menuflex.AddItem(activepage, 0, 1, true)

				tapp.SetFocus(activepage)
			case tcell.KeyEsc, tcell.KeyLeft:
				// pressed escape
				menuflex.
					RemoveItem(activepage).
					RemoveItem(coverbox)
				activatedTable(roottable)
				prelightTable(cattable)
				activateTable(catstable)
				menuflex.AddItem(coverbox, 0, 1, true)
				tapp.SetFocus(catstable)
			}
			return event
		})
		menuflex.
			AddItem(cattable, cattablewidth, 1, false).
			AddItem(coverbox, 0, 1, true)
	})
	catstable.SetSelectedFunc(func(y, x int) {
		menuflex.
			RemoveItem(coverbox).
			RemoveItem(activepage)
		if y == 0 {
			itemname = ""
			prelightTable(catstable)
			activateTable(roottable)
			coverbox.SetText("")
			menuflex.
				AddItem(coverbox, 0, 1, true)
			tapp.SetFocus(roottable)
		} else {
			// itemname = strings.TrimSpace(catstable.GetCell(y, x).Text)
			prelightTable(roottable)
			activatedTable(catstable)
			activateTable(cattable)
			if !(cat == "" || itemname == "") {
				activepage = genPage(cat, itemname, false, app, nil, y)
				menuflex.RemoveItem(coverbox)
				menuflex.AddItem(activepage, 0, 1, true)
			}
			tapp.SetFocus(cattable)
		}
	})
	catstable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyLeft, tcell.KeyEsc:
			menuflex.
				RemoveItem(coverbox).
				RemoveItem(cattable).
				RemoveItem(activepage)
			itemname = ""
			coverbox.SetText("")
			menuflex.
				AddItem(coverbox, 0, 1, true)
			lastTable(cattable)
			prelightTable(catstable)
			activateTable(roottable)
			tapp.SetFocus(roottable)
		case tcell.KeyRight, tcell.KeyTab:
			y, _ := catstable.GetSelection()
			if y == 0 {
				break
			}
			prelightTable(roottable)
			activatedTable(catstable)
			activateTable(cattable)
			if !(cat == "" || itemname == "") {
				activepage = genPage(cat, itemname, false, app, nil, y)
				menuflex.RemoveItem(coverbox)
				menuflex.AddItem(activepage, 0, 1, true)
			}
			tapp.SetFocus(
				cattable)
		}
		return event
	})
	// root is the canvas (the whole current terminal view)
	root := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(titlebar, 1, 0, false).
		AddItem(menuflex, 0, 1, true)

	if e := tapp.SetRoot(root, true).Run(); e != nil {
		panic(e)
	}

	return 0
}

func getMaxWidth(ss []string) (maxwidth int) {
	for _, x := range ss {
		if len(x) > maxwidth {
			maxwidth = len(x)
		}
	}
	return
}

func genMenu(items ...string) (table *tview.Table, menuwidth int) {
	menuwidth = getMaxWidth(items)
	table = tview.NewTable().SetSelectable(true, true)
	table.SetCell(0, 0, tview.NewTableCell("<"))
	for i, x := range items {
		pad := strings.Repeat(" ", menuwidth-len(x))
		table.SetCell(i+1, 0, tview.NewTableCell(" "+pad+x))
	}
	t, l, _, h := table.Box.GetRect()
	menuwidth += 2
	table.Box.SetRect(t, l, menuwidth, h)
	return
}

// This sets a menu to active attributes
func activateTable(table *tview.Table) {
	if table == nil {
		return
	}
	rowcount := table.GetRowCount()
	for i := 0; i < rowcount; i++ {
		table.GetCell(i, 0).
			SetAttributes(tcell.AttrNone).
			SetTextColor(TextColor()).
			SetBackgroundColor(MainColor())
		table.SetSelectedStyle(MainColor(), TextColor(), tcell.AttrBold)
		table.Box.SetBackgroundColor(MainColor())
	}
}

// This sets a menu to activated (it has a selected item active)
func activatedTable(table *tview.Table) {
	if table == nil {
		return
	}
	rowcount := table.GetRowCount()
	for i := 0; i < rowcount; i++ {
		table.GetCell(i, 0).
			SetAttributes(tcell.AttrNone).
			SetTextColor(MainColor()).
			SetBackgroundColor(DimColor())
		table.SetSelectedStyle(DimColor(), MainColor(), tcell.AttrBold)
		table.Box.SetBackgroundColor(DimColor())
	}
}

// This sets a menu to preview (when it is active but not selected yet)
func prelightTable(table *tview.Table) {
	if table == nil {
		return
	}
	rowcount := table.GetRowCount()
	for i := 0; i < rowcount; i++ {
		table.GetCell(i, 0).
			SetAttributes(tcell.AttrNone).
			SetTextColor(DimColor()).
			SetBackgroundColor(PrelightColor())
		table.SetSelectedStyle(PrelightColor(), DimColor(), tcell.AttrBold)
		table.Box.SetBackgroundColor(PrelightColor())
	}
}

// This is just for the one case of the root table with the editor active
func lastTable(table *tview.Table) {
	if table == nil {
		return
	}
	rowcount := table.GetRowCount()
	for i := 0; i < rowcount; i++ {
		table.GetCell(i, 0).
			SetAttributes(tcell.AttrNone).
			SetTextColor(PrelightColor()).
			SetBackgroundColor(BackgroundColor())
		table.SetSelectedStyle(BackgroundColor(), PrelightColor(), tcell.AttrBold)
		table.Box.SetBackgroundColor(BackgroundColor())
	}
}

func MakeConfig(c *App) (out *nine.Config) {
	C := c.Cats
	var configFile string
	var tn, sn, rn bool
	out = &nine.Config{
		ConfigFile:               &configFile,
		AppDataDir:               C.Str("app", "appdatadir"),
		DataDir:                  C.Str("app", "datadir"),
		LogDir:                   C.Str("app", "logdir"),
		LogLevel:                 C.Str("log", "level"),
		Subsystems:               C.Map("log", "subsystem"),
		Network:                  C.Str("p2p", "network"),
		AddPeers:                 C.Tags("p2p", "addpeer"),
		ConnectPeers:             C.Tags("p2p", "connect"),
		MaxPeers:                 C.Int("p2p", "maxpeers"),
		Listeners:                C.Tags("p2p", "listen"),
		DisableListen:            C.Bool("p2p", "nolisten"),
		DisableBanning:           C.Bool("p2p", "disableban"),
		BanDuration:              C.Duration("p2p", "banduration"),
		BanThreshold:             C.Int("p2p", "banthreshold"),
		Whitelists:               C.Tags("p2p", "whitelist"),
		Username:                 C.Str("rpc", "user"),
		Password:                 C.Str("rpc", "pass"),
		ServerUser:               C.Str("rpc", "user"),
		ServerPass:               C.Str("rpc", "pass"),
		LimitUser:                C.Str("limit", "user"),
		LimitPass:                C.Str("limit", "pass"),
		RPCConnect:               C.Str("rpc", "connect"),
		RPCListeners:             C.Tags("rpc", "listen"),
		RPCCert:                  C.Str("tls", "cert"),
		RPCKey:                   C.Str("tls", "key"),
		RPCMaxClients:            C.Int("rpc", "maxclients"),
		RPCMaxWebsockets:         C.Int("rpc", "maxwebsockets"),
		RPCMaxConcurrentReqs:     C.Int("rpc", "maxconcurrentreqs"),
		RPCQuirks:                C.Bool("rpc", "quirks"),
		DisableRPC:               C.Bool("rpc", "disable"),
		NoTLS:                    C.Bool("tls", "disable"),
		DisableDNSSeed:           C.Bool("p2p", "nodns"),
		ExternalIPs:              C.Tags("p2p", "externalips"),
		Proxy:                    C.Str("proxy", "address"),
		ProxyUser:                C.Str("proxy", "user"),
		ProxyPass:                C.Str("proxy", "pass"),
		OnionProxy:               C.Str("proxy", "address"),
		OnionProxyUser:           C.Str("proxy", "user"),
		OnionProxyPass:           C.Str("proxy", "pass"),
		Onion:                    C.Bool("proxy", "tor"),
		TorIsolation:             C.Bool("proxy", "isolation"),
		TestNet3:                 &tn,
		RegressionTest:           &rn,
		SimNet:                   &sn,
		AddCheckpoints:           C.Tags("chain", "addcheckpoints"),
		DisableCheckpoints:       C.Bool("chain", "disablecheckpoints"),
		DbType:                   C.Str("chain", "dbtype"),
		Profile:                  C.Int("app", "profile"),
		CPUProfile:               C.Str("app", "cpuprofile"),
		Upnp:                     C.Bool("app", "upnp"),
		MinRelayTxFee:            C.Float("p2p", "minrelaytxfee"),
		FreeTxRelayLimit:         C.Float("p2p", "freetxrelaylimit"),
		NoRelayPriority:          C.Bool("p2p", "norelaypriority"),
		TrickleInterval:          C.Duration("p2p", "trickleinterval"),
		MaxOrphanTxs:             C.Int("p2p", "maxorphantxs"),
		Algo:                     C.Str("mining", "algo"),
		Generate:                 C.Bool("mining", "generate"),
		GenThreads:               C.Int("mining", "genthreads"),
		MiningAddrs:              C.Tags("mining", "addresses"),
		MinerListener:            C.Str("mining", "listener"),
		MinerPass:                C.Str("mining", "pass"),
		BlockMinSize:             C.Int("block", "minsize"),
		BlockMaxSize:             C.Int("block", "maxsize"),
		BlockMinWeight:           C.Int("block", "minweight"),
		BlockMaxWeight:           C.Int("block", "maxweight"),
		BlockPrioritySize:        C.Int("block", "prioritysize"),
		UserAgentComments:        C.Tags("p2p", "useragentcomments"),
		NoPeerBloomFilters:       C.Bool("p2p", "nobloomfilters"),
		NoCFilters:               C.Bool("p2p", "nocfilters"),
		SigCacheMaxSize:          C.Int("chain", "sigcachemaxsize"),
		BlocksOnly:               C.Bool("p2p", "blocksonly"),
		TxIndex:                  C.Bool("chain", "txindex"),
		AddrIndex:                C.Bool("chain", "addrindex"),
		RelayNonStd:              C.Bool("chain", "relaynonstd"),
		RejectNonStd:             C.Bool("chain", "rejectnonstd"),
		TLSSkipVerify:            C.Bool("tls", "skipverify"),
		Wallet:                   C.Bool("wallet", "enable"),
		NoInitialLoad:            C.Bool("wallet", "noinitialload"),
		WalletPass:               C.Str("wallet", "pass"),
		WalletServer:             C.Str("wallet", "server"),
		CAFile:                   C.Str("tls", "cafile"),
		OneTimeTLSKey:            C.Bool("tls", "onetime"),
		ServerTLS:                C.Bool("tls", "server"),
		LegacyRPCListeners:       C.Tags("rpc", "listen"),
		LegacyRPCMaxClients:      C.Int("rpc", "maxclients"),
		LegacyRPCMaxWebsockets:   C.Int("rpc", "maxwebsockets"),
		ExperimentalRPCListeners: &[]string{},
		State:                    node.StateCfg,
	}
	return
}

func optTagList(s []string) (ss string) {
	if len(ss) > 1 {

		ss = "[<"
		for i, x := range s {
			ss += x
			if i < len(s)-1 {
				ss += ">|<"
			} else {
				ss += ">]"
			}
		}
	}
	return
}

func getCommands(cmds Commands) (s []string) {
	for i := range cmds {
		s = append(s, i)
	}
	sort.Strings(s)
	return
}

func getTokens(cmds Tokens) (s []string) {
	for _, x := range cmds {
		s = append(s, x.Value)
	}
	sort.Strings(s)
	return
}

func Help(args []string, tokens Tokens, app *App) int {
	fmt.Println(app.Name, app.Version(), "-", app.Tagline)
	fmt.Println()
	fmt.Println("help with", app.Name)
	fmt.Println()
	if len(tokens) == 1 {
		// help was invoked
		var tags []string
		for i := range app.Commands {
			tags = append(tags, i)
		}
		sort.Strings(tags)
		for _, x := range tags {
			// if ac := app.Commands[x]; ac.Handler != nil {
			ac := app.Commands[x]
			fmt.Printf("\t%s '%s' %s\n\t\t%s\n\n",
				x, ac.Pattern,
				optTagList(ac.Opts),
				ac.Short)
			// }
		}
	} else {
		// some number of other commands were mentioned
		fmt.Println(
			"showing items mentioned alongside help in commandline:",
			tokens.GetSortedKeys(),
		)
		fmt.Println()
		var tags []string
		for i := range tokens {
			tags = append(tags, i)
		}
		sort.Strings(tags)
		for _, x := range tags {
			if x != "help" {
				fmt.Printf("%s '%s' %s\n\n\t%s\n",
					x, app.Commands[x].Pattern,
					optTagList(app.Commands[x].Opts),
					app.Commands[x].Short)
				fmt.Println("\n", app.Commands[x].Detail)
				fmt.Println()
			}
		}
	}
	return 0
}

func Conf(args []string, tokens Tokens, app *App) int {
	var r int
	for r = 2; r == 2; {
		r = Run(args, tokens, app)
	}
	return r
}

func New(args []string, tokens Tokens, app *App) int {
	fmt.Println("running New", args, getTokens(tokens))
	return 0
}

func Copy(args []string, tokens Tokens, app *App) int {
	fmt.Println("running Copy", args, getTokens(tokens))
	return 0
}

func List(args []string, tokens Tokens, app *App) int {
	if j := validateProxyListeners(app); j != 0 {
		return j
	}
	if _, ok := tokens["wallet"]; ok {
		app.Cats["wallet"]["enable"].Put(true)
	}
	ctl.ListCommands()
	return 0
}

func Ctl(args []string, tokens Tokens, app *App) int {
	// spew.Dump(app.Cats["app"])
	cl.Register.SetAllLevels(*app.Config.LogLevel)
	setAppDataDir(app, "ctl")
	if j := validateProxyListeners(app); j != 0 {
		return j
	}
	if _, ok := tokens["wallet"]; ok {
		*app.Config.Wallet = true
	}
	var i int
	var x string
	for i, x = range args {
		if app.Commands["ctl"].RE.Match([]byte(x)) {
			i++
			break
		}
	}
	ctl.Main(args[i:], app.Config)
	return 0
}

func Node(args []string, tokens Tokens, app *App) int {
	node.StateCfg = app.Config.State
	node.Cfg = app.Config
	cl.Register.SetAllLevels(*app.Config.LogLevel)
	setAppDataDir(app, "node")
	_ = nine.ActiveNetParams //= activenetparams
	if validateWhitelists(app) != 0 ||
		validateProxyListeners(app) != 0 ||
		validatePasswords(app) != 0 ||
		validateRPCCredentials(app) != 0 ||
		validateBlockLimits(app) != 0 ||
		validateUAComments(app) != 0 ||
		validateMiner(app) != 0 ||
		validateCheckpoints(app) != 0 ||
		validateAddresses(app) != 0 ||
		validateDialers(app) != 0 {
		return 1
	}
	// run the node!
	if node.Main(nil) != nil {
		return 1
	}
	return 0
}

func Wallet(args []string, tokens Tokens, app *App) int {
	setAppDataDir(app, "wallet")
	netDir := walletmain.NetworkDir(*app.Config.AppDataDir, app.Config.ActiveNetParams.Params)
	wdb := netDir // + "/wallet.db"
	log <- cl.Debug{"opening wallet:", wdb}
	if !FileExists(wdb) {
		if e := walletmain.CreateWallet(
			app.Config, app.Config.ActiveNetParams, wdb); e != nil {
			panic("could not create wallet " + e.Error())
		}
	} else {
		setAppDataDir(app, "node")
		if e := walletmain.Main(app.Config, app.Config.ActiveNetParams, netDir); e != nil {
			return 1
		}
	}
	return 0
}

func Shell(args []string, tokens Tokens, app *App) int {
	setAppDataDir(app, "node")
	netDir := walletmain.NetworkDir(filepath.Join(*app.Config.DataDir, "wallet"), app.Config.ActiveNetParams.Params)
	wdb := netDir // + "/wallet.db"
	log <- cl.Debug{"opening wallet:", wdb}
	if !FileExists(wdb) {
		if e := walletmain.CreateWallet(
			app.Config, app.Config.ActiveNetParams, wdb); e != nil {
			panic("could not create wallet " + e.Error())
		}
	} else {
		go Node(args, tokens, app)
		if e := walletmain.Main(app.Config, app.Config.ActiveNetParams, netDir); e != nil {
			return 1
		}
	}
	return 0
}

func Test(args []string, tokens Tokens, app *App) int {
	cl.Register.SetAllLevels(*app.Config.LogLevel)
	fmt.Println("running Test", args, getTokens(tokens))
	return 0
}

func Create(args []string, tokens Tokens, app *App) int {
	netDir := walletmain.NetworkDir(filepath.Join(*app.Config.DataDir, "wallet"), app.Config.ActiveNetParams.Params)
	wdb := netDir // + "/wallet.db"
	if !FileExists(wdb) {
		if e := walletmain.CreateWallet(
			app.Config, app.Config.ActiveNetParams, wdb); e != nil {
			panic("could not create wallet " + e.Error())
		}
	} else {
		fmt.Println("wallet already exists in", wdb+"/wallet.db", "refusing to overwrite")
		return 1
	}
	return 0
}

func TestHandler(args []string, tokens Tokens, app *App) int {
	return 0
}

func GUI(args []string, tokens Tokens, app *App) int {
	return 0
}

func Mine(args []string, tokens Tokens, app *App) int {
	return 0
}
func GenCerts(args []string, tokens Tokens, app *App) int {
	return 0
}
func GenCA(args []string, tokens Tokens, app *App) int {
	return 0
}

// UseLogger uses a specified Logger to output package logging info. This should be used in preference to SetLogWriter if the caller is also using log.
func UseLogger(
	logger *cl.SubSystem) {
	Log = logger
	log = Log.Ch
}

func (app *App) Parse(args []string) int {
	// parse commandline
	cmd, tokens := app.ParseCLI(args)
	if cmd == nil {
		cmd = app.Commands["help"]
	}
	// get datadir from cli args if given
	if dd, ok := tokens["datadir"]; ok {
		datadir = &dd.Value
		pwd, _ := os.Getwd()
		*datadir = filepath.Join(pwd, *datadir)
		dd.Value = *datadir
		app.Cats["app"]["datadir"].Value.Put(*datadir)
		DataDir = *datadir
	} else {
		ddd := util.AppDataDir("9", false)
		app.Cats["app"]["datadir"].Put(ddd)
		datadir = &ddd
		DataDir = *datadir
	}
	// for i, x := range app.Cats {
	// 	for j := range x {
	// 		// if i == "app" && j == "datadir" {
	// 		// 	break
	// 		// }
	// 		app.Cats[i][j].Init(app.Cats[i][j])
	// 	}
	// }

	// // set AppDataDir for running as node
	// aa := CleanAndExpandPath(filepath.Join(
	// 	*datadir,
	// 	cmd.Name),
	// 	*datadir)
	// app.Config.AppDataDir, app.Config.LogDir = &aa, &aa

	configFile := CleanAndExpandPath(filepath.Join(
		*datadir, "config"), *datadir)
	// *app.Config.ConfigFile = configFile
	if !FileExists(configFile) {
		if EnsureDir(configFile) {
		}
		fh, err := os.Create(configFile)
		if err != nil {
			panic(err)
		}
		j, e := json.MarshalIndent(app, "", "\t")
		if e != nil {
			panic(e)
		}
		_, err = fmt.Fprint(fh, string(j))
		if err != nil {
			panic(err)
		}
	}
	conf, err := ioutil.ReadFile(configFile)
	if err != nil {
		panic(err)
	}
	e := json.Unmarshal(conf, app)
	if e != nil {
		panic(e)
	}
	// now we can initialise the App
	for i, x := range app.Cats {
		for j := range x {
			temp := app.Cats[i][j]
			temp.App = app
			app.Cats[i][j] = temp
		}
	}
	app.Config = MakeConfig(app)
	app.Config.ActiveNetParams = node.ActiveNetParams

	if app.Config.LogLevel != nil {
		cl.Register.SetAllLevels(*app.Config.LogLevel)
	}
	// run as configured
	r := cmd.Handler(
		args,
		tokens,
		app)
	return r
}

func (app *App) ParseCLI(args []string) (cmd *Command, tokens Tokens) {
	cmd = new(Command)
	// collect set of items in commandline
	if len(args) < 2 {
		fmt.Print("No args given, printing help:\n\n")
		args = append(args, "h")
	}
	commandsFound := make(map[string]int)
	tokens = make(Tokens)
	for _, x := range args[1:] {
		for i, y := range app.Commands {
			if y.RE.MatchString(x) {
				if _, ok := commandsFound[i]; ok {
					tokens[i] = Token{x, *y}
					commandsFound[i]++
					break
				} else {
					tokens[i] = Token{x, *y}
					commandsFound[i] = 1
					break
				}
			}
		}
	}
	var withHandlersNames []string
	withHandlers := make(Commands)
	for i := range commandsFound {
		if app.Commands[i].Handler != nil {
			withHandlers[i] = app.Commands[i]
			withHandlersNames = append(withHandlersNames, i)
		}
	}
	invoked := make(Commands)
	for i, x := range withHandlers {
		invoked[i] = x
	}
	// search the precedents of each in the case of multiple
	// with handlers and delete the one that has another in the
	// list of matching handlers. If one is left we can run it,
	// otherwise return an error.
	var resolved []string
	if len(withHandlersNames) > 1 {
		var common [][]string
		for _, x := range withHandlersNames {
			i := intersection(withHandlersNames, withHandlers[x].Precedent)
			common = append(common, i)
		}
		for _, x := range common {
			for _, y := range x {
				if y != "" {
					resolved = append(resolved, y)
				}
			}
		}
		resolved = uniq(resolved)
		if len(resolved) > 1 {
			withHandlers = make(Commands)
			common = [][]string{}
			withHandlersNames = resolved
			resolved = []string{}
			for _, x := range withHandlersNames {
				withHandlers[x] = app.Commands[x]
			}
			for _, x := range withHandlersNames {
				i := intersection(withHandlersNames, withHandlers[x].Precedent)
				common = append(common, i)
			}
			for _, x := range common {
				for _, y := range x {
					if y != "" {
						resolved = append(resolved, y)
					}
				}
			}
			resolved = uniq(resolved)
		}
	} else if len(withHandlersNames) == 1 {
		resolved = []string{withHandlersNames[0]}
	}
	// fmt.Println(resolved)
	if len(resolved) < 1 {
		err := fmt.Errorf(
			"\nunable to resolve which command to run:\n\tinput: '%s'",
			withHandlersNames)
		fmt.Println(err)
		return nil, tokens
	}
	*cmd = *app.Commands[resolved[0]]
	return cmd, tokens
}

// MinUint32 is a helper function to return the minimum of two uint32s. This avoids a math import and the need to cast to floats.
func MinUint32(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

func isWindows() bool {
	return runtime.GOOS == "windows"
}

// EnsureDir checks a file could be written to a path, creates the directories as needed
func EnsureDir(fileName string) bool {
	dirName := filepath.Dir(fileName)
	if _, serr := os.Stat(dirName); serr != nil {
		merr := os.MkdirAll(dirName, os.ModePerm)
		if merr != nil {
			panic(merr)
		}
		return true
	}
	return false
}

// FileExists reports whether the named file or directory exists.
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

// CleanAndExpandPath expands environment variables and leading ~ in the passed path, cleans the result, and returns it.
func CleanAndExpandPath(path, datadir string) string {
	// Expand initial ~ to OS specific home directory.
	homeDir := filepath.Dir(util.AppDataDir("9", false))
	if strings.HasPrefix(path, "~") {
		return strings.Replace(path, "~", homeDir, 1)

	}
	if strings.HasPrefix(path, "./") {
		// explicitly prefix is this must be a relative path
		pwd, _ := os.Getwd()
		return filepath.Join(pwd, path)
	} else if !strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "\\") {
		if path != datadir {
			return filepath.Join(datadir, path)
		}
	}
	// NOTE: The os.ExpandEnv doesn't work with Windows-style %VARIABLE%, but they variables can still be expanded via POSIX-style $VARIABLE.
	path = filepath.Clean(os.ExpandEnv(path))
	return path
}

// NormalizeAddress returns addr with the passed default port appended if there is not already a port specified.
func NormalizeAddress(addr, defaultPort string) string {
	_, _, err := net.SplitHostPort(addr)
	if err != nil {
		return net.JoinHostPort(addr, defaultPort)
	}
	return addr
}

// NormalizeAddresses returns a new slice with all the passed peer addresses normalized with the given default port, and all duplicates removed.
func NormalizeAddresses(addrs []string, defaultPort string) []string {
	for i, addr := range addrs {
		addrs[i] = NormalizeAddress(addr, defaultPort)
	}
	return RemoveDuplicateAddresses(addrs)
}

// RemoveDuplicateAddresses returns a new slice with all duplicate entries in addrs removed.
func RemoveDuplicateAddresses(addrs []string) []string {
	result := make([]string, 0, len(addrs))
	seen := map[string]struct{}{}
	for _, val := range addrs {
		if _, ok := seen[val]; !ok {
			result = append(result, val)
			seen[val] = struct{}{}
		}
	}
	return result
}

func intersection(a, b []string) (out []string) {
	for _, x := range a {
		for _, y := range b {
			if x == y {
				out = append(out, x)
			}
		}
	}
	return
}

func uniq(elements []string) []string {
	// Use map to record duplicates as we find them.
	encountered := map[string]bool{}
	result := []string{}

	for v := range elements {
		if encountered[elements[v]] == true {
			// Do not add duplicate.
		} else {
			// Record this element as an encountered element.
			encountered[elements[v]] = true
			// Append to result slice.
			result = append(result, elements[v])
		}
	}
	// Return the new slice.
	return result
}

func setAppDataDir(app *App, name string) {
	if app != nil {
		if app.Config != nil {
			if app.Config.AppDataDir == nil {
				app.Config.AppDataDir = new(string)
				// set AppDataDir for running as node
				*app.Config.AppDataDir =
					CleanAndExpandPath(
						filepath.Join(*app.Config.DataDir, name),
						*app.Config.DataDir)
			}
			if app.Config.LogDir == nil {
				app.Config.LogDir = new(string)
				*app.Config.LogDir = *app.Config.AppDataDir
			}
		}
	}
}

func validateWhitelists(app *App) int {
	// Validate any given whitelisted IP addresses and networks.
	if app.Config.Whitelists != nil {
		var ip net.IP

		app.Config.State.ActiveWhitelists =
			make([]*net.IPNet, 0, len(*app.Config.Whitelists))
		for _, addr := range *app.Config.Whitelists {
			_, ipnet, err := net.ParseCIDR(addr)
			if err != nil {
				err = fmt.Errorf("%s '%s'", cl.Ine(), err.Error())
				ip = net.ParseIP(addr)
				if ip == nil {
					str := err.Error() + " %s: the whitelist value of '%s' is invalid"
					err = fmt.Errorf(str, "runNode", addr)
					return 1
				}
				var bits int
				if ip.To4() == nil {
					// IPv6
					bits = 128
				} else {
					bits = 32
				}
				ipnet = &net.IPNet{
					IP:   ip,
					Mask: net.CIDRMask(bits, bits),
				}
			}
			app.Config.State.ActiveWhitelists =
				append(app.Config.State.ActiveWhitelists, ipnet)
		}
	}
	return 0
}

func validateProxyListeners(app *App) int {
	// if proxy is not enabled, empty the proxy field as node sees presence as a
	// on switch
	if app.Config.Proxy != nil {
		*app.Config.Proxy = ""
	}
	// if proxy is enabled or listeners list is empty, or connect peers are set,
	// disable p2p listener
	if app.Config.Proxy != nil ||
		app.Config.ConnectPeers != nil ||
		app.Config.Listeners == nil {
		if app.Config.DisableListen == nil {
			acd := true
			app.Config.DisableListen = &acd
		} else {
			*app.Config.DisableListen = true
		}
	}
	if !*app.Config.DisableListen && len(*app.Config.Listeners) < 1 {
		*app.Config.Listeners = []string{
			net.JoinHostPort("127.0.0.1", node.DefaultPort),
		}
	}
	return 0
}

func validatePasswords(app *App) int {

	// Check to make sure limited and admin users don't have the same username
	if *app.Config.Username != "" && *app.Config.Username == *app.Config.LimitUser {
		str := "%s: --username and --limituser must not specify the same username"
		err := fmt.Errorf(str, "runNode")
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	// Check to make sure limited and admin users don't have the same password
	if *app.Config.Password != "" &&
		*app.Config.Password == *app.Config.LimitPass {
		str := "%s: --password and --limitpass must not specify the same password"
		err := fmt.Errorf(str, "runNode")
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func validateRPCCredentials(app *App) int {
	// The RPC server is disabled if no username or password is provided.
	if (*app.Config.Username == "" || *app.Config.Password == "") &&
		(*app.Config.LimitUser == "" || *app.Config.LimitPass == "") {
		*app.Config.DisableRPC = true
	}
	if *app.Config.DisableRPC {
	}
	if !*app.Config.DisableRPC && len(*app.Config.RPCListeners) == 0 {
		addrs, err := net.LookupHost(node.DefaultRPCListener)
		if err != nil {
			return 1
		}
		*app.Config.RPCListeners = make([]string, 0, len(addrs))
		for _, addr := range addrs {
			addr = net.JoinHostPort(addr, app.Config.ActiveNetParams.RPCPort)
			*app.Config.RPCListeners = append(*app.Config.RPCListeners, addr)
		}
	}
	return 0
}

func validateBlockLimits(app *App) int {
	// Validate the the minrelaytxfee.
	// log <- cl.Debug{"checking min relay tx fee"}
	var err error
	app.Config.State.ActiveMinRelayTxFee, err =
		util.NewAmount(*app.Config.MinRelayTxFee)
	if err != nil {
		str := "%s: invalid minrelaytxfee: %v"
		err := fmt.Errorf(str, "runNode", err)
		fmt.Println(err)
		return 1
	}
	// Limit the block priority and minimum block sizes to max block size.
	*app.Config.BlockPrioritySize = int(MinUint32(
		uint32(*app.Config.BlockPrioritySize),
		uint32(*app.Config.BlockMaxSize)))
	*app.Config.BlockMinSize = int(MinUint32(
		uint32(*app.Config.BlockMinSize),
		uint32(*app.Config.BlockMaxSize)))
	*app.Config.BlockMinWeight = int(MinUint32(
		uint32(*app.Config.BlockMinWeight),
		uint32(*app.Config.BlockMaxWeight)))
	switch {
	// If the max block size isn't set, but the max weight is, then we'll set the limit for the max block size to a safe limit so weight takes precedence.
	case *app.Config.BlockMaxSize == node.DefaultBlockMaxSize &&
		*app.Config.BlockMaxWeight != node.DefaultBlockMaxWeight:
		*app.Config.BlockMaxSize = blockchain.MaxBlockBaseSize - 1000
	// If the max block weight isn't set, but the block size is, then we'll scale the set weight accordingly based on the max block size value.
	case *app.Config.BlockMaxSize != node.DefaultBlockMaxSize &&
		*app.Config.BlockMaxWeight == node.DefaultBlockMaxWeight:
		*app.Config.BlockMaxWeight = *app.Config.BlockMaxSize * blockchain.WitnessScaleFactor
	}
	if *app.Config.RejectNonStd && *app.Config.RelayNonStd {
		fmt.Println("cannot both relay and reject nonstandard transactions")
		return 1
	}
	return 0
}

func validateUAComments(app *App) int {
	// Look for illegal characters in the user agent comments.
	// log <- cl.Debug{"checking user agent comments"}
	if app.Config.UserAgentComments != nil {
		for _, uaComment := range *app.Config.UserAgentComments {
			if strings.ContainsAny(uaComment, "/:()") {
				err := fmt.Errorf("%s: The following characters must not "+
					"appear in user agent comments: '/', ':', '(', ')'",
					"runNode")
				fmt.Fprintln(os.Stderr, err)
				return 1
			}
		}
	}
	return 0
}

func validateMiner(app *App) int {
	// Check mining addresses are valid and saved parsed versions.
	// log <- cl.Debug{"checking mining addresses"}
	if app.Config.MiningAddrs != nil {
		app.Config.State.ActiveMiningAddrs =
			make([]util.Address, 0, len(*app.Config.MiningAddrs))
		if len(*app.Config.MiningAddrs) > 0 {
			for _, strAddr := range *app.Config.MiningAddrs {
				if len(strAddr) > 1 {
					addr, err := util.DecodeAddress(strAddr, app.Config.ActiveNetParams.Params)
					if err != nil {
						str := "%s: mining address '%s' failed to decode: %v"
						err := fmt.Errorf(str, "runNode", strAddr, err)
						fmt.Fprintln(os.Stderr, err)
						return 1
					}
					if !addr.IsForNet(app.Config.ActiveNetParams.Params) {
						str := "%s: mining address '%s' is on the wrong network"
						err := fmt.Errorf(str, "runNode", strAddr)
						fmt.Fprintln(os.Stderr, err)
						return 1
					}
					app.Config.State.ActiveMiningAddrs =
						append(app.Config.State.ActiveMiningAddrs, addr)
				} else {
					*app.Config.MiningAddrs = []string{}
				}
			}
		}
	}
	// Ensure there is at least one mining address when the generate flag
	// is set.
	if (*app.Config.Generate ||
		app.Config.MinerListener != nil) &&
		app.Config.MiningAddrs != nil {
		str := "%s: the generate flag is set, but there are no mining addresses specified "
		err := fmt.Errorf(str, "runNode")
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if *app.Config.MinerPass != "" {
		app.Config.State.ActiveMinerKey = fork.Argon2i([]byte(*app.Config.MinerPass))
	}
	return 0
}

func validateCheckpoints(app *App) int {
	var err error
	// Check the checkpoints for syntax errors.
	// log <- cl.Debug{"checking the checkpoints"}
	if app.Config.AddCheckpoints != nil {
		app.Config.State.AddedCheckpoints, err =
			node.ParseCheckpoints(*app.Config.AddCheckpoints)
		if err != nil {
			str := "%s: Error parsing checkpoints: %v"
			err := fmt.Errorf(str, "runNode", err)
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	return 0
}

func validateDialers(app *App) int {
	// if !*app.Config.Onion && *app.Config.OnionProxy != "" {
	// 	// log <- cl.Error{"cannot enable tor proxy without an address specified"}
	// 	return 1
	// }

	// Tor stream isolation requires either proxy or onion proxy to be set.
	if *app.Config.TorIsolation &&
		app.Config.Proxy == nil {
		str := "%s: Tor stream isolation requires either proxy or onionproxy to be set"
		err := fmt.Errorf(str, "runNode")
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	// Setup dial and DNS resolution (lookup) functions depending on the specified options.  The default is to use the standard net.DialTimeout function as well as the system DNS resolver.  When a proxy is specified, the dial function is set to the proxy specific dial function and the lookup is set to use tor (unless --noonion is specified in which case the system DNS resolver is used).
	// log <- cl.Debug{"setting network dialer and lookup"}
	app.Config.State.Dial = net.DialTimeout
	app.Config.State.Lookup = net.LookupIP
	if app.Config.Proxy != nil {
		fmt.Println("loading proxy")
		// log <- cl.Debug{"we are loading a proxy!"}
		_, _, err := net.SplitHostPort(*app.Config.Proxy)
		if err != nil {
			str := "%s: Proxy address '%s' is invalid: %v"
			err := fmt.Errorf(str, "runNode", *app.Config.Proxy, err)
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		// Tor isolation flag means proxy credentials will be overridden unless
		// there is also an onion proxy configured in which case that one will be overridden.
		torIsolation := false
		if *app.Config.TorIsolation &&
			(app.Config.ProxyUser != nil ||
				app.Config.ProxyPass != nil) {
			torIsolation = true
			// log <- cl.Warn{
			// "Tor isolation set -- overriding specified proxy user credentials"}
		}
		proxy := &socks.Proxy{
			Addr:         *app.Config.Proxy,
			Username:     *app.Config.ProxyUser,
			Password:     *app.Config.ProxyPass,
			TorIsolation: torIsolation,
		}
		app.Config.State.Dial = proxy.DialTimeout
		// Treat the proxy as tor and perform DNS resolution through it unless the --noonion flag is set or there is an onion-specific proxy configured.
		if *app.Config.Onion &&
			*app.Config.OnionProxy != "" {
			app.Config.State.Lookup = func(host string) ([]net.IP, error) {
				return connmgr.TorLookupIP(host, *app.Config.Proxy)
			}
		}
	}
	// Setup onion address dial function depending on the specified options. The default is to use the same dial function selected above.  However, when an onion-specific proxy is specified, the onion address dial function is set to use the onion-specific proxy while leaving the normal dial function as selected above.  This allows .onion address traffic to be routed through a different proxy than normal traffic.
	// log <- cl.Debug{"setting up tor proxy if enabled"}
	if app.Config.OnionProxy != nil {
		_, _, err := net.SplitHostPort(*app.Config.OnionProxy)
		if err != nil {
			str := "%s: Onion proxy address '%s' is invalid: %v"
			err := fmt.Errorf(str, "runNode", *app.Config.OnionProxy, err)
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		// Tor isolation flag means onion proxy credentials will be overriddenode.
		if *app.Config.TorIsolation &&
			(*app.Config.OnionProxyUser != "" || *app.Config.OnionProxyPass != "") {
			// log <- cl.Warn{
			// "Tor isolation set - overriding specified onionproxy user credentials "}
		}
		app.Config.State.Oniondial =
			func(network, addr string, timeout time.Duration) (net.Conn, error) {
				proxy := &socks.Proxy{
					Addr:         *app.Config.OnionProxy,
					Username:     *app.Config.OnionProxyUser,
					Password:     *app.Config.OnionProxyPass,
					TorIsolation: *app.Config.TorIsolation,
				}
				return proxy.DialTimeout(network, addr, timeout)
			}
		// When configured in bridge mode (both --onion and --proxy are configured), it means that the proxy configured by --proxy is not a tor proxy, so override the DNS resolution to use the onion-specific proxy.
		if *app.Config.Proxy != "" {
			app.Config.State.Lookup = func(host string) ([]net.IP, error) {
				return connmgr.TorLookupIP(host, *app.Config.OnionProxy)
			}
		}
	} else {
		app.Config.State.Oniondial = app.Config.State.Dial
	}
	// Specifying --noonion means the onion address dial function results in an error.
	if !*app.Config.Onion {
		app.Config.State.Oniondial = func(a, b string, t time.Duration) (net.Conn, error) {
			return nil, errors.New("tor has been disabled")
		}
	}
	return 0
}

func validateAddresses(app *App) int {
	// TODO: simplify this to a boolean and one slice for config fercryinoutloud
	if app.Config.AddPeers != nil && app.Config.ConnectPeers != nil {
		fmt.Println("ERROR:", cl.Ine(),
			"cannot have addpeers at the same time as connectpeers")
		return 1
	}
	// Add default port to all rpc listener addresses if needed and remove duplicate addresses.
	// log <- cl.Debug{"checking rpc listener addresses"}
	*app.Config.RPCListeners =
		node.NormalizeAddresses(*app.Config.RPCListeners,
			app.Config.ActiveNetParams.RPCPort)
	// Add default port to all listener addresses if needed and remove duplicate addresses.
	if app.Config.Listeners != nil {
		*app.Config.Listeners =
			node.NormalizeAddresses(*app.Config.Listeners,
				app.Config.ActiveNetParams.DefaultPort)
	}
	// Add default port to all added peer addresses if needed and remove duplicate addresses.
	if app.Config.AddPeers != nil {
		*app.Config.AddPeers =
			node.NormalizeAddresses(*app.Config.AddPeers,
				app.Config.ActiveNetParams.DefaultPort)
	}
	if app.Config.ConnectPeers != nil {
		*app.Config.ConnectPeers =
			node.NormalizeAddresses(*app.Config.ConnectPeers,
				app.Config.ActiveNetParams.DefaultPort)
	}
	// --onionproxy and not --onion are contradictory (TODO: this is kinda stupid hm? switch *and* toggle by presence of flag value, one should be enough)
	return 0
}
