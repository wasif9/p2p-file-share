package main

import (
	"image/color"

	"gioui.org/app"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

func PopupMessage(message string) {
	go func() {
		popupWindow := new(app.Window)
		popupWindow.Option(app.Title("P2P File Share"))
		popupWindow.Option(app.Size(unit.Dp(400), unit.Dp(200)))

		thPopup := material.NewTheme()
		var popupOps op.Ops

		// Run the popup window
		for {
			popupEvent := popupWindow.Event()
			switch popupEvent := popupEvent.(type) {
			case app.FrameEvent:
				popupGtx := app.NewContext(&popupOps, popupEvent)

				// Display the message in the popup window
				layout.Flex{Axis: layout.Vertical}.Layout(popupGtx,
					layout.Flexed(1, func(popupGtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal}.Layout(popupGtx,
							layout.Flexed(1, func(popupGtx layout.Context) layout.Dimensions {
								return layout.Center.Layout(popupGtx, func(popupGtx layout.Context) layout.Dimensions {
									msg := material.H3(thPopup, message)
									msg.Font.Weight = font.Bold
									msg.Color = color.NRGBA{R: 255, A: 255}
									msg.TextSize = unit.Sp(20)
									return msg.Layout(popupGtx)
								})
							}),
						)
					}),
				)

				popupEvent.Frame(popupGtx.Ops)

			case app.DestroyEvent:
				return
			}
		}
	}()
}
