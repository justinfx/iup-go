package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"net/http"
	"unsafe"

	"github.com/gen2brain/iup-go/iup"
)

func main() {
	iup.Open()
	defer iup.Close()

	labelImage := iup.Label("").SetHandle("label").SetAttribute("IMAGE", "xkcd")
	button := iup.Button("Random XKCD").SetHandle("button")

	vbox := iup.Vbox(labelImage, button).SetAttributes(map[string]string{
		"ALIGNMENT": "ACENTER",
		"GAP":       "10",
		"MARGIN":    "10x10",
	})

	dlg := iup.Dialog(vbox).SetAttribute("TITLE", "PostMessage")
	dlg.SetAttribute("RESIZE", "NO")
	dlg.SetHandle("dlg")

	labelImage.SetCallback("POSTMESSAGE_CB", iup.PostMessageFunc(messageCb))
	button.SetCallback("ACTION", iup.ActionFunc(buttonCb))

	iup.ShowXY(dlg, iup.CENTER, iup.CENTER)

	buttonCb(button)

	iup.MainLoop()
}

func messageCb(ih iup.Ihandle, s string, i int, f float64, p unsafe.Pointer) int {
	b := unsafe.Slice((*byte)(p), i)
	img, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		log.Fatalln(err)
	}

	iup.ImageFromImage(img).SetHandle("xkcd")

	iup.GetHandle("label").SetAttribute("TIP", s)
	iup.GetHandle("label").SetAttribute("IMAGE", "xkcd")
	iup.GetHandle("dlg").SetAttribute("RASTERSIZE", fmt.Sprintf("%dx%d", img.Bounds().Dx()+20, img.Bounds().Dy()+80))

	iup.GetHandle("button").SetAttribute("ACTIVE", "YES")

	iup.Refresh(iup.GetHandle("dlg"))
	iup.ShowXY(iup.GetHandle("dlg"), iup.CENTER, iup.CENTER)

	return iup.DEFAULT
}

func buttonCb(ih iup.Ihandle) int {
	ih.SetAttribute("ACTIVE", "NO")

	go func() {
		res, err := http.Get("https://random-xkcd-img.herokuapp.com/")
		if err != nil {
			log.Println(err)
		}
		defer res.Body.Close()

		var ret map[string]string
		err = json.NewDecoder(res.Body).Decode(&ret)
		if err != nil {
			log.Println(err)
		}

		img, err := http.Get(ret["url"])
		if err != nil {
			log.Println(err)
		}
		defer img.Body.Close()

		b, err := ioutil.ReadAll(img.Body)
		if err != nil {
			log.Println(err)
		}

		iup.PostMessage(iup.GetHandle("label"), ret["title"], len(b), 1.0, unsafe.Pointer(&b[0]))
	}()

	return iup.DEFAULT
}