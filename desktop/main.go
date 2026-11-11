package main

import (
	"log"
	"path/filepath"
	"time"

	"github.com/justdoswift/justlearn/desktop/internal/store"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/updater"
	"github.com/egoist/mygo/plugins/updater/native"
	"github.com/egoist/mygo/ui"
)

func main() {
	loadDotEnv()
	mygo.Use(native.Plugin)
	mygo.App.WhenReady(func() {
		devMode := mygo.IsDev()
		startedAt := time.Now()
		dir, err := mygo.App.Path(mygo.PathUserData)
		if err != nil {
			log.Fatal(err)
		}
		st, err := store.Open(filepath.Join(dir, "justlearn.db"))
		if err != nil {
			log.Fatal(err)
		}
		if err := st.EnsureSeed(); err != nil {
			log.Fatal(err)
		}
		appearance := st.GetSetting("appearance", "system")
		switch appearance {
		case "light", "dark":
		default:
			appearance = "system"
		}
		mygo.Theme.SetSource(mygo.ThemeSource(appearance))
		a := newApp(st)
		a.appearance = appearance
		a.systemDark = mygo.Theme.IsDark()
		switch appearance {
		case "light":
			a.dark = false
		case "dark":
			a.dark = true
		default:
			a.dark = a.systemDark
		}
		a.version = mygo.App.Version()
		a.win = mygo.NewWindow(mygo.WindowOptions{
			Title:                "JustLearn",
			Width:                1240,
			Height:               820,
			MinWidth:             1120,
			MinHeight:            580,
			StateKey:             "main",
			BackgroundColor:      "light-dark(#f3f1eb, #171815)",
			TitleBarStyle:        mygo.TitleBarHidden,
			TrafficLightPosition: &mygo.Point{X: 14, Y: 14},
			Vibrancy:             mygo.VibrancyGlass,
			Hidden:               devMode,
			Content:              ui.View(a.view),
		})
		if devMode {
			// `mygo dev` relaunches the app on every edit. Start hidden so a
			// rebuild never takes focus from whatever the user is doing; a
			// Dock click or the menu opens the window when they want it.
			showWindow := func() {
				if !a.win.IsVisible() {
					a.win.Show()
				}
			}
			mygo.App.OnActivate(func(bool) {
				// A relaunch can activate the app itself; only a later
				// user activation (Dock click) should open the window.
				if time.Since(startedAt) > 2*time.Second {
					showWindow()
				}
			})
		}
		if cfgKey(st) == "" {
			a.openSettingsAt(1)
		} else {
			a.focusDraft = true
		}
		mygo.Theme.OnUpdated(func() {
			a.win.Update(func() { a.syncSystemAppearance() })
		})
		mygo.App.SetMenu(mygo.NewMenu([]*mygo.MenuItem{
			{Label: "JustLearn", Submenu: []*mygo.MenuItem{
				{Role: mygo.RoleAbout},
				updater.MenuItem(),
				mygo.Separator(),
				{Role: mygo.RoleHide},
				{Role: mygo.RoleHideOthers},
				{Role: mygo.RoleUnhide},
				mygo.Separator(),
				{Role: mygo.RoleQuit},
			}},
			{Label: "文件", Submenu: []*mygo.MenuItem{
				{Label: "显示窗口", Accelerator: "CmdOrCtrl+O", Click: func(*mygo.MenuItem, *mygo.Window) {
					a.win.Show()
				}},
				{Label: "新建学习路线", Accelerator: "CmdOrCtrl+N", Click: func(*mygo.MenuItem, *mygo.Window) {
					a.win.Update(func() { a.newRoute() })
				}},
				{Label: "回到首页", Accelerator: "CmdOrCtrl+Shift+H", Click: func(*mygo.MenuItem, *mygo.Window) {
					a.win.Update(func() { a.goHome() })
				}},
				mygo.Separator(),
				{Label: "设置…", Accelerator: "CmdOrCtrl+,", Click: func(*mygo.MenuItem, *mygo.Window) {
					a.win.Update(func() { a.openSettings() })
				}},
			}},
			{Role: mygo.RoleEditMenu},
			{Role: mygo.RoleWindowMenu},
		}))
		a.win.OnClose(func(_ *mygo.CloseEvent) {
			a.cancelSend()
			_ = st.Close()
		})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
