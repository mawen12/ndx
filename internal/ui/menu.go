package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type MenuOption struct {
	Text     string
	Selected func()
}

type Menu struct {
	*tview.Box

	disabled bool

	options []*MenuOption

	open bool

	field string

	fieldWidth int

	fieldBackgroundColor tcell.Color

	fieldTextColor tcell.Color

	list *tview.List

	selected func(text string, index int)

	dragging bool
}

func NewMenu() *Menu {
	list := tview.NewList()
	list.ShowSecondaryText(false).
		SetMainTextStyle(tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tview.Styles.PrimitiveBackgroundColor)).
		SetSelectedStyle(tcell.StyleDefault.Background(tview.Styles.PrimaryTextColor).Foreground(tview.Styles.PrimitiveBackgroundColor)).
		SetHighlightFullLine(true).
		SetBackgroundColor(tcell.ColorBlue)

	box := tview.NewBox()

	m := &Menu{
		Box:                  box,
		list:                 list,
		fieldBackgroundColor: tview.Styles.ContrastBackgroundColor,
		fieldTextColor:       tview.Styles.PrimaryTextColor,
	}

	return m
}

func (m *Menu) AddOption(text string, selected func()) *Menu {
	m.options = append(m.options, &MenuOption{Text: text, Selected: selected})
	m.list.AddItem(text, "", 0, nil)
	return m
}

func (m *Menu) SetOptions(texts []string, selected func(text string, index int)) *Menu {
	m.list.Clear()
	m.options = nil

	for _, text := range texts {
		m.AddOption(text, nil)
	}
	m.selected = selected
	return m
}

func (m *Menu) GetOptionCount() int {
	return len(m.options)
}

func (m *Menu) RemoveOption(index int) *Menu {
	m.options = append(m.options[:index], m.options[index+1:]...)
	m.list.RemoveItem(index)
	return m
}

func (m *Menu) SetField(field string) *Menu {
	m.field = field
	return m
}

func (m *Menu) GetField() string {
	return m.field
}

func (m *Menu) SetFieldWidth(width int) *Menu {
	m.fieldWidth = width
	return m
}

func (m *Menu) GetFieldWidth() int {
	return m.fieldWidth
}

func (m *Menu) SetFieldBackgroundColor(color tcell.Color) *Menu {
	m.fieldBackgroundColor = color
	return m
}

func (m *Menu) GetFieldBackgroundColor() tcell.Color {
	return m.fieldBackgroundColor
}

func (m *Menu) SetFieldTextColor(color tcell.Color) *Menu {
	m.fieldTextColor = color
	return m
}

func (m *Menu) GetFieldTextColor() tcell.Color {
	return m.fieldTextColor
}

// func (m *Menu) SetNext(p PrimitiveFunc) *Menu {
// 	m.next = p
// 	return m
// }

// func (m *Menu) SetPrev(p PrimitiveFunc) *Menu {
// 	m.prev = p
// 	return m
// }

func (m *Menu) SetSelectedFunc(handler func(text string, index int)) *Menu {
	m.selected = handler
	return m
}

func (m *Menu) Draw(screen tcell.Screen) {
	m.Box.DrawForSubclass(screen, m)

	x, y, width, height := m.GetInnerRect()
	rightLimit := x + width
	if height < 1 || rightLimit <= x {
		return
	}
	useStyleTags, _ := m.list.GetUseStyleTags()

	// Draw field
	fieldWidth := m.fieldWidth
	if fieldWidth == 0 {
		fieldWidth = tview.TaggedStringWidth(m.field)
	}
	fieldStyle := tcell.StyleDefault.Background(m.fieldBackgroundColor)
	if m.HasFocus() && !m.open {
		fieldStyle = fieldStyle.Background(m.fieldTextColor)
	}
	for index := 0; index < fieldWidth; index++ { // print with style
		screen.SetContent(x+index, y, ' ', nil, fieldStyle)
	}

	maxWidth := 0
	for _, option := range m.options {
		str := option.Text
		if !useStyleTags {
			str = tview.Escape(str)
		}
		strWidth := tview.TaggedStringWidth(str)
		if strWidth > maxWidth {
			maxWidth = strWidth
		}
	}

	color := m.fieldTextColor
	text := m.field
	if m.HasFocus() && !m.open {
		color = m.fieldBackgroundColor
	}
	tview.Print(screen, text, x, y, fieldWidth, tview.AlignLeft, color) // print content

	// Draw optiosn list
	if m.HasFocus() && m.open {
		lx := x
		ly := y + 1
		lwidth := maxWidth
		lheight := len(m.options)
		swidth, sheight := screen.Size()
		if lx+lwidth >= swidth {
			lx = max(swidth-lwidth, 0)
		}

		if ly+lheight >= sheight && ly-2 > lheight-ly {
			ly = y - height
			if ly < 0 {
				ly = 0
			}
		}

		if ly+lheight >= sheight {
			lheight = sheight - ly
		}

		m.list.SetRect(lx, ly, lwidth, lheight)
		m.list.Draw(screen)
	}
}

func (m *Menu) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return m.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		if m.disabled {
			return
		}

		key := event.Key()
		switch key {
		case tcell.KeyEnter, tcell.KeyDown:
			if !m.open {
				m.openList(setFocus)
			} else if handler := m.list.InputHandler(); handler != nil {
				handler(event, setFocus)
			}
		case tcell.KeyEscape:
			if m.open {
				m.CloseList(setFocus)
			}
		// case tcell.KeyTab:
		// 	if p.open {
		// 		p.CloseList(setFocus)
		// 	}

		// 	if p.next != nil {
		// 		if np := p.next(); np != nil {
		// 			setFocus(np)
		// 		}
		// 	}
		// case tcell.KeyBacktab:
		// 	if p.open {
		// 		p.CloseList(setFocus)
		// 	}
		// 	if p.prev != nil {
		// 		if pp := p.prev(); pp != nil {
		// 			setFocus(pp)
		// 		}
		// 	}
		case tcell.KeyUp:
			if handler := m.list.InputHandler(); m.open && handler != nil {
				handler(event, setFocus)
			}
		default:
			m.CloseList(setFocus)
		}
	})
}

func (m *Menu) openList(setFocus func(tview.Primitive)) {
	if m.open {
		return
	}

	m.open = true

	m.list.
		SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
			if m.dragging {
				return
			}

			m.CloseList(setFocus)

			currentOption := m.options[index]
			if m.selected != nil {
				m.selected(currentOption.Text, index)
			}
			if currentOption.Selected != nil {
				currentOption.Selected()
			}
		}).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch key := event.Key(); key {
			case tcell.KeyDown, tcell.KeyUp, tcell.KeyPgUp, tcell.KeyHome, tcell.KeyEnd, tcell.KeyEnter:
				break
			case tcell.KeyEscape:
				m.CloseList(setFocus)
			default:
				return nil
			}
			return event
		})

	setFocus(m.list)
}

func (m *Menu) CloseList(setFocus func(tview.Primitive)) {
	m.open = false
	if m.list.HasFocus() {
		setFocus(m)
	}
}

func (m *Menu) IsOpen() bool {
	return m.open
}

func (m *Menu) Focus(delegate func(p tview.Primitive)) {
	if m.open {
		delegate(m.list)
	} else {
		m.Box.Focus(delegate)
	}
}

func (m *Menu) HasFocus() bool {
	if m.open {
		return m.list.HasFocus()
	}
	return m.Box.HasFocus()
}

func (m *Menu) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return m.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		x, y := event.Position()
		inRect := m.InInnerRect(x, y)
		if !m.open && !inRect {
			return m.InRect(x, y), nil
		}

		if m.open {
			capture = m
		}

		switch action {
		case tview.MouseLeftDown:
			consumed = m.open || inRect
			capture = m
			if !m.open {
				m.openList(setFocus)
				m.dragging = true
			} else if consumed, _ := m.list.MouseHandler()(tview.MouseLeftClick, event, setFocus); !consumed {
				m.CloseList(setFocus)
			}
		case tview.MouseMove:
			if m.dragging {
				m.list.MouseHandler()(tview.MouseLeftClick, event, setFocus)
				consumed = true
			}
		case tview.MouseLeftUp:
			if m.dragging {
				m.dragging = false
				m.list.MouseHandler()(tview.MouseLeftClick, event, setFocus)
				consumed = true
			}
		}

		return
	})
}
