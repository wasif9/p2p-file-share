package main

import (
	"image/color"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

func MainLayout(gtx layout.Context, th *material.Theme, dnUI *DownloadUI, upUI *UploadUI, prUI *ProgressUI, tabButtons []widget.Clickable) layout.Dimensions {
	// Layout the tabs
	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		// Left side: tab buttons
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				tab_Btn(th, &tabButtons[0], "Download", DownloadTab),
				tab_Btn(th, &tabButtons[1], "Upload", UploadTab),
				tab_Btn(th, &tabButtons[2], "Progress", ProgressTab),
			)
		}),
		// Right side: whichever tab is selected
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			switch TabSelected {
			case DownloadTab:
				return dnUI.DownloadLayout(gtx, th, prUI)
			case UploadTab:
				return upUI.UploadLayout(gtx, th)
			case ProgressTab:
				return prUI.ProgressLayout(gtx, th)
			}
			return layout.Dimensions{}
		}),
	)
}

func tab_Btn(th *material.Theme, button *widget.Clickable, title string, tab int) layout.FlexChild {
	return layout.Flexed(0.7, func(gtx layout.Context) layout.Dimensions {
		// Padding
		inset := layout.Inset{Top: 10, Right: 20, Bottom: 10, Left: 20}

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
				TabSelected = tab
			}

			return btn.Layout(gtx)
		})
	})
}
