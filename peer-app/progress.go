package main

import (
	"image/color"
	"log"
	"sync"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// Download represents a file being downloaded
type Download struct {
	Hash     string
	Name     string
	Progress float32
	Checked  widget.Bool
	Shown    bool
	mu       sync.Mutex
}

// UI struct manages the downloading files list
type ProgressUI struct {
	list      widget.List
	files     []*Download
	deleteBtn widget.Clickable
}

func (ui *ProgressUI) ProgressLayout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical, Alignment: layout.End}.Layout(gtx,
		// Title
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						// Padding
						inset := layout.Inset{
							Top:    10,
							Bottom: 10,
						}

						return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							text := material.Body1(th, "Downloading Files")
							text.TextSize = unit.Sp(20)
							return text.Layout(gtx)
						})
					})
				}),
			)
		}),

		// Delete Buttons
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.End}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					// Padding
					inset := layout.Inset{
						Right: 20,
					}

					return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(th, &ui.deleteBtn, "Delete")
						btn.Background = color.NRGBA{R: 255, G: 0, B: 0, A: 255} // Red color for delete
						gtx.Constraints.Min.X = gtx.Dp(100)
						gtx.Constraints.Min.Y = gtx.Dp(50)

						// Action when click delete button
						if ui.deleteBtn.Clicked(gtx) {
							ui.deleteCheckedFiles()
						}

						return btn.Layout(gtx)
					})
				}),
			)
		}),

		// Download List
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.list.Layout(gtx, len(ui.files), func(gtx layout.Context, i int) layout.Dimensions {
				if ui.files[i].Shown {
					// Padding
					inset := layout.Inset{
						Top:    10,
						Bottom: 10,
						Left:   20,
						Right:  20,
					}

					return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return ui.layoutDownloadItem(gtx, th, ui.files[i])
					})
				} else {
					return layout.Dimensions{}
				}
			})
		}),
	)
}

// Layout of downloading files
func (ui *ProgressUI) layoutDownloadItem(gtx layout.Context, th *material.Theme, file *Download) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		// Checkbox
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return material.CheckBox(th, &file.Checked, "").Layout(gtx)
		}),

		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				// File name
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					label := material.Body1(th, file.Name)
					label.TextSize = unit.Sp(16)
					label.Color = color.NRGBA{A: 255}
					return label.Layout(gtx)
				}),
				// Progress Bar
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					progress := file.GetProgress()
					bar := material.ProgressBar(th, progress)
					return bar.Layout(gtx)
				}),
			)
		}),
	)
}

// Add new download
func (ui *ProgressUI) AddDownload(name string, hash string) *Download {
	if ui.IsInDownload(name, hash) {
		return nil
	}

	log.Println("New file " + name + "starts to download")
	newFile := Download{Hash: hash, Name: name, Progress: 0, Shown: true}

	ui.files = append(ui.files, &newFile)

	return ui.files[len(ui.files)-1]
}

// Check if is in download list
func (ui *ProgressUI) IsInDownload(name string, hash string) bool {
	for _, file := range ui.files {
		if file == nil {
			continue
		}
		// If file is in the download list
		progress := file.GetProgress()
		if hash == file.Hash && progress != 1 {
			return true
		}
	}

	return false
}

// Remove all checked downloads
func (ui *ProgressUI) deleteCheckedFiles() {
	for i := range ui.files {
		// Let checked file not to be shown in the list
		if ui.files[i].Checked.Value {
			ui.files[i].Shown = false
			log.Println(ui.files[i].Name + " is removed from the list")
		}
	}
}

func (file *Download) UpdateProgress(totalChunks int) {
	file.mu.Lock()
	defer file.mu.Unlock()
	file.Progress += 1 / float32(totalChunks)
}

func (file *Download) GetProgress() float32 {
	file.mu.Lock()
	defer file.mu.Unlock()
	return file.Progress
}
