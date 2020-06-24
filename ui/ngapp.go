package ui

import (
	"fmt"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"github.com/zhenggao2/ngapp/nrgrid"
)

// NgApp contains the UI implementation.
type NgApp struct {
	// enable debugging?
	debug bool

	// zap logger
	logger *zap.Logger

	// widgets
	MainWin   *widgets.QMainWindow
	tabWidget *widgets.QTabWidget
	logEdit   *widgets.QTextEdit

	// actions
	exitAction        *widgets.QAction
	nrGridAction *widgets.QAction
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
	app.exitAction = widgets.NewQAction2("Exit", app.MainWin)
	app.exitAction.ConnectTriggered(func(checked bool) { app.MainWin.Close() })

	app.nrGridAction = widgets.NewQAction2("NR Resource Grid", app.MainWin)
	app.nrGridAction.ConnectTriggered(app.onExecNgGrid)

	app.enableDebugAction = widgets.NewQAction2("Enable Debug", app.MainWin)
	app.enableDebugAction.SetCheckable(true)
	app.enableDebugAction.SetChecked(false)
	app.enableDebugAction.ConnectTriggered(func(checked bool) { app.debug = checked })

	app.aboutAction = widgets.NewQAction2("About", app.MainWin)
	app.aboutAction.ConnectTriggered(func(checked bool) {
		info := "<h1>ngapp</h1><p>ngapp is a collection of useful applets for 4G and 5G NPO(Network Planning&Optimization).</p>" +
			"<p>Author: <a href=mailto: zhengwei.gao@yahoo.com>zhengwei.gao@yahoo.com</a></p>" +
			"<p>Blog: <a href=\"http: //blog.csdn.net/jeffyko\">http: //blog.csdn.net/jeffyko</a></p>"
		widgets.QMessageBox_Information(app.MainWin, "About ngapp", info, widgets.QMessageBox__Ok, widgets.QMessageBox__NoButton)
	})
	app.aboutQtAction = widgets.NewQAction2("About Qt", app.MainWin)
	app.aboutQtAction.ConnectTriggered(func(checked bool) { widgets.QMessageBox_AboutQt(app.MainWin, "About Qt") })
}

func (app *NgApp) createMenus() {
	app.fileMenu = app.MainWin.MenuBar().AddMenu2("File")
	app.fileMenu.QWidget_PTR().AddAction(app.exitAction)

	app.lteMenu = app.MainWin.MenuBar().AddMenu2("LTE")

	app.nrMenu = app.MainWin.MenuBar().AddMenu2("NR")
	app.nrMenu.QWidget_PTR().AddAction(app.nrGridAction)

	app.miscMenu = app.MainWin.MenuBar().AddMenu2("Misc")

	app.optionsMenu = app.MainWin.MenuBar().AddMenu2("Options")
	app.optionsMenu.QWidget_PTR().AddAction(app.enableDebugAction)

	app.helpMenu = app.MainWin.MenuBar().AddMenu2("Help")
	app.helpMenu.QWidget_PTR().AddAction(app.aboutAction)
	app.helpMenu.QWidget_PTR().AddAction(app.aboutQtAction)
}

func (app *NgApp) onExecNgGrid(checked bool) {
	ui := new(nrgrid.NrGridUi)
	ui.Debug = app.debug
	ui.Logger = app.logger
	ui.LogEdit = app.logEdit
	ui.Args = nil
	ui.InitUi()
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
		debug: false,
		MainWin:     mainWin,
		tabWidget:   tabWidget,
		logEdit:     logEdit,
	}

	app.logger = logger.WithOptions(zap.Hooks(func(entry zapcore.Entry) error {
		app.logEdit.Append(fmt.Sprintf("<b>[%v]</b> : %v : %v :<br>%v", entry.Level.CapitalString(), entry.Time.Format("2006-01-02 15:04:05.999"), entry.Caller.TrimmedPath(), entry.Message))
		if entry.Level >= zap.ErrorLevel {
			app.logEdit.Append(entry.Stack)
		}
		return nil
	}))

	app.createActions()
	app.createMenus()

	return app
}