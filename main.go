package main

import (
	"fmt"
	"github.com/gobuffalo/packr"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/zevst/homemade_screenshotter/app"
	"github.com/zevst/homemade_screenshotter/ui"
	"log"
	"os"
)

var (
	// variables passed as build flags
	UploadUrl string
	TmpFolder string
	AccessKey string
)

func init() {
	cfg := app.GetConfig()
	cfg.UploadUrl = UploadUrl
	cfg.TmpFolder = TmpFolder
	cfg.AccessKey = AccessKey
	fs := app.GetFS()
	fs.Templates = packr.NewBox("./templates")
}

func main() {
	application, err := gtk.ApplicationNew("hmsc.msz.client", glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		log.Fatal(err)
	}
	_, err = application.Connect("activate", func() {
		gtk.Init(nil)

		win := ui.SetupWindow(application, "Homemade Screenshotter")
		box := ui.SetupBox(gtk.ORIENTATION_VERTICAL)
		win.Add(box)

		// both textboxes bust be added in a scrollable container, otherwise they force whole window to expand
		historyView, historyViewWrapper := ui.SetupLabel()
		box.PackStart(historyViewWrapper, false, true, 0)

		logView, logViewWrapper := ui.SetupLabel()
		box.Add(logViewWrapper)

		btn := ui.SetupBtn("Upload", func() {
			doUpload(historyView, logView)
		})
		box.PackEnd(btn, false, false, 0)

		// Recursively show all widgets contained in this window.
		win.ShowAll()
		// Begin executing the GTK main loop.  This blocks until gtk.MainQuit() is run.
		gtk.Main()
	})
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(application.Run(os.Args))
}

// doUpload performs uploading text/img to server; triggers by click on button
func doUpload(historyView *gtk.Label, logView *gtk.Label) {
	clipboard := ui.GetClipboard()
	if textContent, textErr := clipboard.WaitForText(); textErr == nil {
		go func() {
			ui.SetTextGlib(logView, "upload Text ... ")
			if fileUrl, err := app.SendTextToServer(textContent, app.GetConfig().UploadUrl); err != nil {
				ui.SetTextGlib(logView, fmt.Sprintf("Error uploading TEXT.\nDetails: \"%s\"", err.Error()))
				return
			} else {
				ui.PrependMarkupGlib(historyView, fmt.Sprintf("TXT: <a href=\"%s\">%s</a>", fileUrl, fileUrl))
				ui.SetClipboardTextGlib(clipboard, fileUrl)
				ui.SetTextGlib(logView, "TEXT upload OK, URL copied to clipboard")
			}
		}()
	} else if imageContent, imageErr := clipboard.WaitForImage(); imageErr == nil {
		go func() {
			ui.SetTextGlib(logView, "upload Image ... ")
			if fileUrl, err := app.SendImageToServer(imageContent, app.GetConfig().UploadUrl); err != nil {
				ui.SetTextGlib(logView, fmt.Sprintf("Error uploading IMAGE.\nDetails: \"%s\"", err.Error()))
				return
			} else {
				ui.PrependMarkupGlib(historyView, fmt.Sprintf("IMG: <a href=\"%s\">%s</a>", fileUrl, fileUrl))
				ui.SetClipboardTextGlib(clipboard, fileUrl)
				ui.SetTextGlib(logView, "IMAGE upload OK, URL copied to clipboard")
			}
		}()
	} else {
		logView.SetText("upload nothing - clipboard content undefined ")
	}
}
