package main

import (
	"fmt"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NgApp contains the UI implementation.
type NgApp struct {
	// enable debugging?
	enableDebug bool

	// zap logger
	logger *zap.Logger

	// widgets
	mainWin   *widgets.QMainWindow
	tabWidget *widgets.QTabWidget
	logEdit   *widgets.QTextEdit

	// actions
	exitAction        *widgets.QAction
	enableDebugAction *widgets.QAction
	aboutAction       *widgets.QAction
	aboutQtAction     *widgets.QAction

	// menus
	fileMenu    *widgets.QMenu
	lteMenu     *widgets.QMenu
	nrMenu      *widgets.QMenu
	miscMenu    *widgets.QMenu
	optionsMenu *widgets.QMenu
	helpMenu    *widgets.QMenu
}

func (app *NgApp) createActions() {
	app.exitAction = widgets.NewQAction2("Exit", app.mainWin)
	app.exitAction.ConnectTriggered(func(checked bool) { app.mainWin.Close() })

	app.enableDebugAction = widgets.NewQAction2("Enable Debug", app.mainWin)
	app.enableDebugAction.SetCheckable(true)
	app.enableDebugAction.SetChecked(false)
	app.enableDebugAction.ConnectTriggered(func(checked bool) { app.enableDebug = checked })

	app.aboutAction = widgets.NewQAction2("About", app.mainWin)
	app.aboutAction.ConnectTriggered(func(checked bool) {
		info := "<h1>ngapp</h1><p>ngapp is a collection of useful applets for 4G and 5G NPO(Network Planning&Optimization).</p>" +
			"<p>Author: <a href=mailto: zhengwei.gao@yahoo.com>zhengwei.gao@yahoo.com</a></p>" +
			"<p>Blog: <a href=\"http: //blog.csdn.net/jeffyko\">http: //blog.csdn.net/jeffyko</a></p>"
		widgets.QMessageBox_Information(app.mainWin, "About ngapp", info, widgets.QMessageBox__Ok, widgets.QMessageBox__NoButton)
	})
	app.aboutQtAction = widgets.NewQAction2("About Qt", app.mainWin)
	app.aboutQtAction.ConnectTriggered(func(checked bool) { widgets.QMessageBox_AboutQt(app.mainWin, "About Qt") })
}

func (app *NgApp) createMenus() {
	app.fileMenu = app.mainWin.MenuBar().AddMenu2("File")
	app.fileMenu.QWidget_PTR().AddAction(app.exitAction)

	app.lteMenu = app.mainWin.MenuBar().AddMenu2("LTE")
	app.nrMenu = app.mainWin.MenuBar().AddMenu2("NR")
	app.miscMenu = app.mainWin.MenuBar().AddMenu2("Misc")

	app.optionsMenu = app.mainWin.MenuBar().AddMenu2("Options")
	app.optionsMenu.QWidget_PTR().AddAction(app.enableDebugAction)

	app.helpMenu = app.mainWin.MenuBar().AddMenu2("Help")
	app.helpMenu.QWidget_PTR().AddAction(app.aboutAction)
	app.helpMenu.QWidget_PTR().AddAction(app.aboutQtAction)
}

// NewNgApp initialize widgets/actions/menus and returns a pointer to NgApp.
func NewNgApp(logger *zap.Logger) *NgApp {
	mainWin := widgets.NewQMainWindow(nil, core.Qt__Widget)
	tabWidget := widgets.NewQTabWidget(nil)
	tabWidget.SetTabsClosable(true)
	logEdit := widgets.NewQTextEdit(nil)
	tabWidget.AddTab(logEdit, "log")

	mainWin.SetCentralWidget(tabWidget)
	mainWin.SetWindowTitle("ngapp")
	mainWin.SetWindowFlags(mainWin.WindowFlags() | core.Qt__WindowMinMaxButtonsHint)
	mainWin.SetWindowState(mainWin.WindowState() | core.Qt__WindowMaximized)
	tabWidget.ConnectTabCloseRequested(func(index int) {
		if index == 0 {
			return
		}
		tabWidget.RemoveTab(index)
	})

	app := &NgApp{
		enableDebug: false,
		mainWin:     mainWin,
		tabWidget:   tabWidget,
		logEdit:     logEdit,
	}

	app.logger = logger.WithOptions(zap.Hooks(func(entry zapcore.Entry) error {
		app.logEdit.Append(fmt.Sprintf("<b>[%v]</b> : %v : %v : %s", entry.Level.CapitalString(), entry.Time.Format("2006-01-02 15:04:05.999"), entry.Caller.TrimmedPath(), entry.Message))
		if entry.Level >= zap.ErrorLevel {
			app.logEdit.Append(entry.Stack)
		}
		return nil
	}))

	app.createActions()
	app.createMenus()

	return app
}
