package main

import (
	"fmt"
	"os"
	"time"
	"github.com/therecipe/qt/widgets"
	"github.com/zhenggao2/ngapp/utils"
)

func main() {
	logger := utils.NewZapLogger(fmt.Sprintf("./logs/ngapp_%v.log", time.Now().Format("20060102_150406")))

	app := widgets.NewQApplication(len(os.Args), os.Args)
	ngapp := NewNgApp(logger)
	ngapp.MainWin.Show()

	/*
	// defer for recover when panic happens
	defer func() {
		if err := recover(); err != nil {
			ngapp.logEdit.Append(fmt.Sprintf("<font color=red><b>[Panic recover]</b></font>: %v.", err))
			app.Exec()
		}
	}()
	 */
	app.Exec()
}
