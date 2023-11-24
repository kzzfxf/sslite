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

package ui

import (
	termui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type MainUI struct {
}

// ShowMainUI
func ShowMainUI() (err error) {
	if err = termui.Init(); err != nil {
		return
	}
	defer termui.Close()

	table1 := widgets.NewTable()
	table1.Rows = [][]string{
		{"header1", "header2", "header3"},
		{"你好吗", "Go-lang is so cool", "Im working on Ruby"},
		{"2016", "10", "11"},
	}
	table1.TextStyle = termui.NewStyle(termui.ColorWhite)
	table1.SetRect(0, 0, 60, 7)
	termui.Render(table1)

	for e := range termui.PollEvents() {
		if e.Type == termui.KeyboardEvent && e.ID == "<C-c>" {
			break
		}
	}
	return
}
