package ui

import (
	"context"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/mum4k/termdash/widgets/textinput"
	"github.com/raneamri/nextop/io"
	"github.com/raneamri/nextop/queries"
	"github.com/raneamri/nextop/types"
	"github.com/raneamri/nextop/utility"

	_ "github.com/go-sql-driver/mysql"
)

/*
Workload:

	Minimal (dynamic ticker-less display, refresh on input)
*/

/*
Format:

	widget-1 (top-left): feedback log
	widget-2 (bottom-left): configurated instances
	widget-3 (top-right): input prompt
	widget-4 (bottom-right): settings
*/
func DisplayConfigs() {
	t, err := tcell.New()
	defer t.Close()

	ctx, cancel := context.WithCancel(context.Background())

	var (
		dbms    string
		dsn     string
		name    string
		log_msg string
	)

	/*
		widget-1
	*/
	errlog, _ := text.New()
	errlog.Write("\n   Awaiting submission.", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))

	/*
		widget-2
	*/
	instlog, _ := text.New()
	instanceDisplay(ctx, instlog)

	/*
		widget-3
	*/
	dbmsin, err := textinput.New(
		textinput.Label("DBMS ", cell.Bold(), cell.FgColor(cell.ColorNumber(33))),
		textinput.TextColor(cell.ColorWhite),
		textinput.MaxWidthCells(45),
		textinput.ExclusiveKeyboardOnFocus(),
		textinput.Border(linestyle.Light),
		textinput.BorderColor(cell.Color(cell.ColorAqua)),
		textinput.PlaceHolder(" <mysql, oracle ...>"),
	)
	dsnin, err := textinput.New(
		textinput.Label("DSN  ", cell.Bold(), cell.FgColor(cell.ColorNumber(33))),
		textinput.TextColor(cell.ColorWhite),
		textinput.MaxWidthCells(45),
		textinput.ExclusiveKeyboardOnFocus(),
		textinput.Border(linestyle.Light),
		textinput.BorderColor(cell.Color(cell.ColorAqua)),
		textinput.PlaceHolder(" <user:pass@tcp(host:port)/name>"),
	)
	namein, err := textinput.New(
		textinput.Label("NAME ", cell.Bold(), cell.FgColor(cell.ColorNumber(33))),
		textinput.TextColor(cell.ColorWhite),
		textinput.MaxWidthCells(45),
		textinput.ExclusiveKeyboardOnFocus(),
		textinput.Border(linestyle.Light),
		textinput.BorderColor(cell.Color(cell.ColorAqua)),
		textinput.PlaceHolder(" <recommended>"),
	)

	/*
		widget-4
	*/
	settings_txt, _ := text.New()
	settings_headers := []string{"refresh-rate",
		"errlog-refresh-rate",
		"default-group",
		"max-processlist-len",
		"max-errlog-len",
		"case-sensitive-filters"}
	dir, _ := os.Getwd()
	settings_txt.Write("\n   "+dir+"/nextop.conf\n", text.WriteCellOpts(cell.Bold()))
	for _, header := range settings_headers {
		settings_txt.Write("\n   "+header, text.WriteCellOpts(cell.FgColor(cell.ColorBlue)))
		settings_txt.Write(": " + io.FetchSetting(header))
	}

	cont, err := container.New(
		t,
		container.ID("configs_display"),
		container.Border(linestyle.Light),
		container.BorderTitle("CONFIGS (ESC to go back, ENTER to submit)"),
		container.BorderColor(cell.ColorGray),
		container.FocusedColor(cell.ColorWhite),
		container.KeyFocusNext(keyboard.KeyTab),
		container.KeyFocusGroupsNext(keyboard.KeyArrowDown, 1),
		container.KeyFocusGroupsPrevious(keyboard.KeyArrowUp, 1),
		container.SplitHorizontal(
			container.Top(
				container.SplitVertical(
					container.Left(
						container.Border(linestyle.Light),
						container.BorderTitle("Status"),
						container.PlaceWidget(errlog),
					),
					container.Right(
						container.Border(linestyle.Light),
						container.BorderTitle("Input"),
						container.SplitHorizontal(
							container.Top(
								container.PlaceWidget(dbmsin),
							),
							container.Bottom(
								container.SplitHorizontal(
									container.Top(
										container.PlaceWidget(dsnin),
									),
									container.Bottom(
										container.PlaceWidget(namein),
									),
									container.SplitPercent(50),
								),
							),
							container.SplitPercent(33),
						),
					),
				),
			),
			container.Bottom(
				container.Border(linestyle.Light),
				container.BorderTitle("Configs"),
				container.SplitVertical(
					container.Left(
						container.Border(linestyle.Light),
						container.BorderTitle("Instances"),
						container.PlaceWidget(instlog),
					),
					container.Right(
						container.Border(linestyle.Light),
						container.BorderTitle("Settings"),
						container.PlaceWidget(settings_txt),
					),
					container.SplitPercent(60),
				),
			),
			container.SplitPercent(40),
		),
	)
	if err != nil {
		panic(err)
	}

	/*
		Config has its own keyboard subscriber
	*/
	keyninreader := func(k *terminalapi.Keyboard) {
		elapsed := time.Since(LastInputTime)
		ratelim, _ := strconv.Atoi(io.FetchSetting("rate-limiter"))
		if elapsed < time.Duration(ratelim)*time.Millisecond {
			return
		}
		LastInputTime = time.Now()

		switch k.Key {
		case keyboard.KeyEnter:
			/*
				Validate data
			*/
			errlog.Reset()

			dbms = dbmsin.ReadAndClear()
			if utility.Dbmsstr(dbms) == -1 {
				log_msg = "\n   Error: Unknown DBMS: " + dbms + "\n"
				errlog.Reset()
				errlog.Write(log_msg, text.WriteCellOpts(cell.FgColor(cell.ColorRed)))
				dsn = dsnin.ReadAndClear()
				name = namein.ReadAndClear()
				return
			}

			dsn = dsnin.ReadAndClear()
			if string(dsn) == "" {
				log_msg = "\n   Error: Blank DSN is invalid."
				errlog.Reset()
				errlog.Write(log_msg, text.WriteCellOpts(cell.FgColor(cell.ColorRed)))
				name = namein.ReadAndClear()
				return
			}

			name = namein.ReadAndClear()
			if name == "" {
				log_msg = "\n   Warning: Blank connection name!"
				errlog.Write(log_msg, text.WriteCellOpts(cell.FgColor(cell.ColorYellow)))
				name = "<unnamed>"
			}

			var inst types.Instance
			inst.DBMS = utility.Dbmsstr(dbms)
			inst.DSN = []byte(dsn)
			inst.ConnName = name

			if !queries.Ping(inst) {
				errlog.Reset()
				errlog.Write("\n   Authenticating...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))
				time.Sleep(1 * time.Second)
				errlog.Reset()
				log_msg = "\n   Error: Invalid DSN. Connection closed."
				errlog.Write(log_msg, text.WriteCellOpts(cell.FgColor(cell.ColorRed)))
				return
			} else {
				errlog.Reset()
				errlog.Write("\n   Authenticating...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))
				time.Sleep(1 * time.Second)
				errlog.Reset()
				log_msg = "\n   Success! Connection established."
				errlog.Write(log_msg, text.WriteCellOpts(cell.FgColor(cell.ColorGreen)))
			}

			inst.Driver, err = queries.Connect(inst)
			if err == nil {
				ActiveConns = append(ActiveConns, inst.ConnName)
				if len(ActiveConns) == 1 {
					CurrConn = ActiveConns[0]
				}
			}

			/*
				Keep if valid and sync to prevent dupes
			*/
			Instances[inst.ConnName] = inst
			io.SyncConfig(Instances)

			/*
				Update instances display
			*/
			instanceDisplay(ctx, instlog)

		case keyboard.KeyCtrlD:
			io.SyncConfig(Instances)
			for _, inst := range Instances {
				if inst.Driver == nil {
					var err error
					inst.Driver, err = queries.Connect(inst)
					if err == nil {
						ActiveConns = append(ActiveConns, inst.ConnName)
					}
				}
			}
			cancel()
		case '?':
			if len(ActiveConns) > 0 {
				State = types.MENU
				cancel()
			} else {
				errlog.Reset()
				log_msg = "\n   Please make sure to have a minimum of one connection online\n   before changing views."
				errlog.Write(log_msg, text.WriteCellOpts(cell.FgColor(cell.ColorRed)))
			}
		case keyboard.KeyEsc:
			if len(ActiveConns) > 0 {
				State = Laststate
				cancel()
			} else {
				errlog.Reset()
				log_msg = "\n   Please make sure to have a minimum of one connection online\n   before changing views."
				errlog.Write(log_msg, text.WriteCellOpts(cell.FgColor(cell.ColorRed)))
			}
		}
	}

	if err := termdash.Run(ctx, t, cont, termdash.KeyboardSubscriber(keyninreader), termdash.RedrawInterval(Interval)); err != nil {
		panic(err)
	}
}

func instanceDisplay(ctx context.Context,
	instlog *text.Text) {

	instlog.Reset()
	for _, inst := range Instances {
		instlog.Write("\n   dbms", text.WriteCellOpts(cell.FgColor(cell.ColorBlue)))
		instlog.Write(": " + utility.Strdbms(inst.DBMS))
		instlog.Write("   dsn", text.WriteCellOpts(cell.FgColor(cell.ColorBlue)))
		instlog.Write(": " + string((inst.DSN)))
		instlog.Write("   conn-name", text.WriteCellOpts(cell.FgColor(cell.ColorBlue)))
		instlog.Write(": " + string((inst.ConnName)))
		if inst.Driver != nil {
			instlog.Write(" ONLINE", text.WriteCellOpts(cell.FgColor(cell.ColorGreen)))
		} else {
			instlog.Write(" OFFLINE", text.WriteCellOpts(cell.FgColor(cell.ColorRed)))
		}
	}
}
