package main

import (
	"image"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type SelectUI struct {
	title       string
	list        widget.List
	directories []os.DirEntry
	dirPath     string
	refreshBtn  widget.Clickable
	backBtn     widget.Clickable
	dirBtns     []widget.Clickable
	confirmBtn  widget.Clickable
	done        bool

	mu sync.RWMutex
}

// LoadDirs loads the list of directories in the current dirPath
func (selectDir *SelectUI) LoadDirs() {
	selectDir.mu.Lock()
	defer selectDir.mu.Unlock()

	log.Println("Current Dir =", selectDir.dirPath)

	files, err := os.ReadDir(selectDir.dirPath)
	if err != nil {
		log.Println("Error reading current directory:", err)
		PopupMessage("Cannot read current directory!")
		selectDir.directories = nil
		return
	}
	// Get all directories
	var directories []os.DirEntry
	for _, file := range files {
		if file.IsDir() {
			directories = append(directories, file)
		}
	}
	// Sort all directories
	sort.Slice(directories, func(i, j int) bool {
		return directories[i].Name() < directories[j].Name()
	})
	selectDir.directories = directories
	selectDir.dirBtns = make([]widget.Clickable, len(directories))
}

// Main layout for the Select Dir UI
func (selectDir *SelectUI) SelectLayout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical, Alignment: layout.End}.Layout(gtx,
		// Title label
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						// Padding
						inset := layout.Inset{Top: 10, Bottom: 10}
						return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							title := material.Body1(th, selectDir.title)
							title.Font.Weight = font.Bold
							title.Color = color.NRGBA{R: 125, G: 150, B: 78, A: 255}
							title.TextSize = unit.Sp(30)
							return title.Layout(gtx)
						})
					})
				}),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				// "Back" button
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					inset := layout.Inset{Top: 10, Bottom: 10, Left: 20}
					return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(th, &selectDir.backBtn, "🢨")
						btn.TextSize = 25
						if selectDir.backBtn.Clicked(gtx) {
							selectDir.navigateUp()
						}
						return btn.Layout(gtx)
					})
				}),
				// Current directory
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					inset := layout.Inset{Top: 10, Bottom: 10, Left: 20, Right: 20}
					return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						label := material.Label(th, th.TextSize, selectDir.dirPath)
						label.Alignment = text.Start
						return label.Layout(gtx)
					})
				}),
				// "Refresh" button
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					inset := layout.Inset{Top: 10, Bottom: 10, Right: 20}
					return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(th, &selectDir.refreshBtn, "⟳")
						btn.TextSize = 25
						if selectDir.refreshBtn.Clicked(gtx) {
							selectDir.LoadDirs()
						}
						return btn.Layout(gtx)
					})
				}),
			)
		}),
		// Dir List
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return selectDir.list.List.Layout(gtx, len(selectDir.directories), func(gtx layout.Context, i int) layout.Dimensions {
				if i >= len(selectDir.directories) {
					return layout.Dimensions{}
				}
				return selectDir.fileLayout(gtx, th, i)
			})
		}),
		// Divider
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			inset := layout.Inset{Top: 10, Bottom: 10}
			return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				width := gtx.Constraints.Max.X
				thickness := gtx.Dp(3)
				rect := clip.Rect{Max: image.Point{X: width, Y: thickness}}.Op()
				paint.FillShape(gtx.Ops, color.NRGBA{R: 210, G: 210, B: 210, A: 255}, rect)
				return layout.Dimensions{Size: image.Point{X: width, Y: thickness}}
			})
		}),
		// Confirm button
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.End}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					inset := layout.Inset{Bottom: 10, Left: 20, Right: 20}
					return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(th, &selectDir.confirmBtn, "Confirm")
						if selectDir.confirmBtn.Clicked(gtx) {
							selectDir.done = true
						}
						return btn.Layout(gtx)
					})
				}),
			)
		}),
	)
}

func (selectDir *SelectUI) fileLayout(gtx layout.Context, th *material.Theme, i int) layout.Dimensions {
	file := selectDir.directories[i]
	btn := &selectDir.dirBtns[i]
	label := file.Name()

	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			inset := layout.Inset{Top: 8, Bottom: 8, Left: 20, Right: 20}
			return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				fileBtn := material.Button(th, btn, label)

				// Blue for folders
				fileBtn.Background = color.NRGBA{R: 100, G: 150, B: 255, A: 255}

				if btn.Clicked(gtx) {
					selectDir.navigateTo(file.Name())
				}
				return fileBtn.Layout(gtx)
			})
		}),
	)
}

// NavigateTo goes deeper into a subdirectory
func (selectDir *SelectUI) navigateTo(dir string) {
	selectDir.dirPath = filepath.Join(selectDir.dirPath, dir)
	selectDir.LoadDirs()
}

// NavigateUp moves up one directory
func (selectDir *SelectUI) navigateUp() {
	selectDir.dirPath = filepath.Dir(selectDir.dirPath)
	selectDir.LoadDirs()
}
