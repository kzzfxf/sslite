// Copyright 2023 kzzfxf
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gui

import (
	"sync/atomic"

	termui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type MainWindow struct {
	gui      *termui.Block
	bridges  *widgets.Table
	tunnels  *widgets.Table
	children []termui.Drawable
	closed   uint32
}

// NewMainWindow
func NewMainWindow() (win *MainWindow, err error) {
	if err = termui.Init(); err != nil {
		return
	}

	win = &MainWindow{
		gui:     termui.NewBlock(),
		bridges: widgets.NewTable(),
		tunnels: widgets.NewTable(),
	}
	// Init UI
	win.initUI()
	win.adaptUI()

	win.children = append(win.children,
		win.bridges,
		win.tunnels,
	)

	return
}

// initUI
func (win *MainWindow) initUI() {
	// Tunnels table
	win.tunnels.Title = "Tunnels"
	win.tunnels.TextStyle = termui.NewStyle(termui.ColorWhite)
	win.tunnels.RowStyles[0] = termui.NewStyle(termui.ColorWhite, termui.StyleClear.Bg, termui.ModifierBold)
	win.tunnels.Rows = [][]string{
		{"Name", "UP↑", "DOWN↓"}, // Headers
	}
	// Bridges table
	win.bridges.Title = "Bridges"
	win.bridges.TextStyle = termui.NewStyle(termui.ColorWhite)
	win.bridges.RowStyles[0] = termui.NewStyle(termui.ColorWhite, termui.StyleClear.Bg, termui.ModifierBold)
	win.bridges.Rows = [][]string{
		{"InBound", "Rule", "OutBound", "RealOutBound", "Tunnel"}, // Headers
	}
}

// adaptUI
func (win *MainWindow) adaptUI() {
	width, height := termui.TerminalDimensions()
	win.gui.SetRect(-1, -1, width+1, height+1)
	wr := win.gui.GetRect()

	win.tunnels.ColumnWidths = []int{20, 8, 8} // 40 - 4
	win.tunnels.SetRect(wr.Max.X-1-40, 0, wr.Max.X-1, wr.Max.Y-1)
	tr := win.tunnels.GetRect()

	win.bridges.SetRect(0, 0, tr.Min.X-1, wr.Max.Y-1)
}

// Show
func (win *MainWindow) Show(fn func()) {
	go func() {
		defer func() {
			win.Close()
			if fn != nil {
				fn()
			}
		}()

		win.Render()

		// Events
		for e := range termui.PollEvents() {
			if e.Type == termui.ResizeEvent {
				win.adaptUI()
				win.Render()
			}
			if e.Type == termui.KeyboardEvent && e.ID == "<C-c>" {
				break
			}
		}
	}()
}

// UpdateTunnelsTable
func (win *MainWindow) UpdateTunnelsTable(rows [][]string) {
	win.tunnels.Rows = append(win.tunnels.Rows[0:1], rows...)
}

// UpdateBridgesTable
func (win *MainWindow) UpdateBridgesTable(rows [][]string) {
	win.bridges.Rows = append(win.bridges.Rows[0:1], rows...)
}

// Render
func (win *MainWindow) Render() {
	termui.Render(win.gui)
	termui.Render(win.children...)
}

// Close
func (win *MainWindow) Close() {
	if !atomic.CompareAndSwapUint32(&win.closed, 0, 1) {
		return
	}
	termui.Close()
}
