package main

import (
    "image/color"
    "log"

    "gioui.org/app"
    "gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/op"
    "gioui.org/widget"
    "gioui.org/widget/material"
)

const (
	DownloadTab = iota		// 0
	UploadTab				// 1
	ProgressTab				// 2
)

var tabSelected = DownloadTab

func main() {
    go func() {
        w := new(app.Window)
		w.Option(app.Title("P2P File Share"))
		w.Option(app.Size(unit.Dp(800), unit.Dp(600)))

        th := material.NewTheme()
		var ops op.Ops
		
		tabButtons := make([]widget.Clickable, 3)

		// DownloadUI instance
		dnUI := &DownloadUI{
			list: widget.List{
				List: layout.List{
					Axis: layout.Vertical,
				},
			},
		}

		// Event
		for {
			e := w.Event()
			switch e := e.(type) {
			case app.DestroyEvent:
				log.Fatal(e.Err)
			case app.FrameEvent:
				gtx := app.NewContext(&ops, e)

				// Layout the tabs
				layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					// Tabs on the left
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							tab_Btn(th, gtx, &tabButtons[0], "Download", DownloadTab),
							tab_Btn(th, gtx, &tabButtons[1], "Upload", UploadTab),
							tab_Btn(th, gtx, &tabButtons[2], "Progress", ProgressTab),
						)
					}),

					// Right content of the selected tab
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if tabSelected == DownloadTab {
							return dnUI.DownloadLayout(gtx, th)
						} else if tabSelected == UploadTab {
							return material.Body1(th, "Upload").Layout(gtx)
						} else {
							return material.Body1(th, "Download Progress").Layout(gtx)
						}
					}),
					
				)
				e.Frame(gtx.Ops)
			}
		}
    }()
    app.Main()
}

func tab_Btn(th *material.Theme, gtx layout.Context, button *widget.Clickable, title string, tab int) layout.FlexChild {
	return layout.Flexed(0.7, func(gtx layout.Context) layout.Dimensions {
		// Padding
		inset := layout.Inset{
			Top:    10,
			Right:  20,
			Bottom: 10,
			Left:   20,
		}

		// Apply the padding and layout the button
		return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			btn := material.Button(th, button, title)

			// Button property
			gtx.Constraints.Min.X = gtx.Dp(120)
			gtx.Constraints.Min.Y = gtx.Dp(80)
			btn.TextSize = unit.Sp(20)
			btn.Color = color.NRGBA{R: 0, G: 0, B: 0, A: 255}
			btn.Background = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	
			// Click event
			if (*button).Clicked(gtx) {
				tabSelected = tab
			}
	
			return btn.Layout(gtx)
		})
	})
}