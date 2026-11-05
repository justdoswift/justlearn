package main

import (
	"log"
	"path/filepath"

	"github.com/justdoswift/justlearn/desktop/internal/store"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/updater"
	"github.com/egoist/mygo/plugins/updater/native"
	"github.com/egoist/mygo/ui"
)

func main() {
	loadDotEnv()
	mygo.Use(native.Plugin)
	mygo.Theme.SetSource(mygo.ThemeDark)
	mygo.App.WhenReady(func() {
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
		a := newApp(st)
		a.win = mygo.NewWindow(mygo.WindowOptions{
			Title:                "JustLearn",
			Width:                1240,
			Height:               820,
			MinWidth:             1120,
			MinHeight:            580,
			StateKey:             "main",
			BackgroundColor:      "#1E1E1E",
			TitleBarStyle:        mygo.TitleBarHidden,
			TrafficLightPosition: &mygo.Point{X: 14, Y: 14},
			Content:              ui.View(a.view),
		})
		if cfgKey(st) == "" {
			a.settingsOpen = true
		}
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
