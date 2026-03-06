package histogram

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/mawen12/ndx/pkg/buf"
	"github.com/rivo/tview"
)

var qblocks = []rune{
	' ', '▗', '▖', '▄', '▝', '▐', '▞', '▟',
	'▘', '▚', '▌', '▙', '▀', '▜', '▛', '█',
}

type fieldData struct {
	dots                [][]bool
	dataBinsInChartBar  int
	chartBarWidth       int
	dotYScale           int
	max                 int
	yScale              int
	effectiveWidthDots  int
	effectiveWidthRunes int
	selScaleDots        [][]bool
	selScaleOffset      int
	cursorVal           int
	selectedValsSum     int
}

type histogramScale struct {
	from, to           int
	numDataBins        int
	dataBinsInChartBar int
	chartBarWidth      int
}

type Histogram struct {
	*tview.Box

	from, to               int
	binSize                int
	data                   map[int64]int
	getXMarks              func(from, to int, numChars int) []int
	xFormat                func(v int) string
	formatCursor           func(from int, to *int, width int) string
	snapDataBinsInChartDot func(dataBinsInChartBar int) int
	selected               func(from, to int)
	curMarks               []int
	cursor                 int
	selectionStart         int
	fldData                *fieldData

	externalCursor        int
	externalCursorVisible bool
}

func NewHistogram() *Histogram {
	h := Histogram{
		Box: tview.NewBox(),
	}

	return &h
}

func (h *Histogram) SetRange(from, to int) *Histogram {
	h.from = from
	h.to = to

	h.cursor = h.to - h.binSize*h.getDataBinsInChartBar()
	h.cursor = h.alignCursor(h.cursor, false)
	h.selectionStart = 0

	return h
}

func (h *Histogram) alignCursor(cursor int, isCeiling bool) int {
	divisor := h.getDataBinsInChartBar() * h.binSize
	cursor -= h.from
	remainder := cursor % divisor
	cursor -= remainder

	if isCeiling && remainder > 0 {
		cursor += divisor
	}
	cursor += h.from
	return cursor
}

func (h *Histogram) SetBinSize(binSize int) *Histogram {
	h.binSize = binSize
	return h
}

func (h *Histogram) SetData(data map[int64]int) *Histogram {
	slog.Info("Histogram set data", "data", data)
	h.data = data
	return h
}

func (h *Histogram) SetXFormat(xFormat func(v int) string) *Histogram {
	h.xFormat = xFormat
	return h
}

func (h *Histogram) SetCursorFormat(formatCursor func(from int, to *int, width int) string) *Histogram {
	h.formatCursor = formatCursor
	return h
}

func (h *Histogram) SetDataBinsSnapper(snapDataBinsInChartDot func(dataBinsInChartBar int) int) *Histogram {
	h.snapDataBinsInChartDot = snapDataBinsInChartDot
	return h
}

func (h *Histogram) SetXMarker(getXMarks func(from, to, numCharts int) []int) *Histogram {
	h.getXMarks = getXMarks
	return h
}

func (h *Histogram) SetExternalCursor(externalCursor int) *Histogram {
	h.externalCursor = externalCursor
	return h
}

func (h *Histogram) HideExternalCursor() *Histogram {
	h.externalCursorVisible = false
	return h
}

func (h *Histogram) ShowExternalCursor() *Histogram {
	h.externalCursorVisible = true
	return h
}

func (h *Histogram) Draw(screen tcell.Screen) {
	h.Box.DrawForSubclass(screen, h)
	x, y, width, height := h.GetInnerRect()

	fldMarginLeft := 0

	fldWidth := (width - fldMarginLeft) * 2
	fldHeight := (height - 1) * 2
	fldData := h.getFieldData(fldWidth, fldHeight)
	if fldData == nil {
		return
	}

	h.fldData = fldData
	h.cursor = h.alignCursor(h.cursor, false)

	fldMarginLeft = (width - fldData.effectiveWidthRunes) / 2

	lines := h.fldDataToLines(fldData.dots)

	for lineY, line := range lines {
		tview.Print(screen, line, x+fldMarginLeft, y+lineY, width-fldMarginLeft, tview.AlignLeft, tcell.ColorLightGray)
	}

	maxLabel := fmt.Sprintf("%d", fldData.yScale)
	maxLabelOffset := fldMarginLeft - len(maxLabel) - 1
	printDot := true
	if maxLabelOffset < 0 {
		maxLabelOffset = 0
		printDot = false
	}
	if printDot {
		maxLabel += "[yellow]▀[-]"
	}
	tview.Print(screen, maxLabel, x+maxLabelOffset, y, width-maxLabelOffset, tview.AlignLeft, tcell.ColorWhite)

	h.curMarks = h.getXMarks(h.from, h.to, width-fldMarginLeft)

	sb := strings.Builder{}
	numRunes := 0

	ruleBuffer := buf.RuneBuffer{}

	for _, mark := range h.curMarks {
		markStr := h.xFormat(mark)
		dotCoord := h.valToCoord(mark)
		charCoord := dotCoord / 2
		remaining := charCoord - numRunes
		for i := 0; i < remaining; i++ {
			sb.WriteRune(' ')
			numRunes++
		}

		sb.WriteString("[yellow]")
		if dotCoord&0x01 != 0 {
			sb.WriteRune('▝')
		} else {
			sb.WriteRune('▘')
		}
		sb.WriteString("[-] ")
		numRunes += 2
		sb.WriteString(markStr)
		numRunes += len(clearViewFormatting(markStr))
	}

	ruleStr := ruleBuffer.String()
	tview.Print(screen, ruleStr, x+fldMarginLeft, y+height-1, width-fldMarginLeft, tview.AlignLeft, tcell.ColorWhite)
	ruleBuffer.WriteAt(x+fldMarginLeft, clearViewFormatting(ruleStr))

	if h.HasFocus() {
		selScaleLines := h.fldDataToLines(fldData.selScaleDots)
		line := selScaleLines[0]
		lineLen := len(fldData.selScaleDots[0]) / 2

		var selMark string
		if !h.IsSelectionActive() {
			selMark = fmt.Sprintf("[%s]", h.formatCursor(h.cursor, nil, h.binSize))
		} else {
			selStart, selEnd := h.GetSelection()
			selMark = fmt.Sprintf("[%s]", h.formatCursor(selStart, &selEnd, h.binSize))
		}

		leftOffset := fldMarginLeft + fldData.selScaleOffset/2

		tview.Print(screen, line, x+leftOffset, y+height-1, width-leftOffset, tview.AlignLeft, tcell.ColorLightGreen)

		selMarkOffset := leftOffset + lineLen + 1
		freeSpaceRight := width - (selMarkOffset + len(selMark))
		if freeSpaceRight < 0 {
			selMarkOffset += freeSpaceRight
		} else {
			selMarkOffset--
			selMark = " " + selMark
		}
		tview.Print(screen, selMark, x+selMarkOffset, y+height-1, width-selMarkOffset, tview.AlignLeft, tcell.ColorLightGreen)

		var valToPrint string
		if !h.IsSelectionActive() {
			valToPrint = fmt.Sprintf("(%d)", fldData.cursorVal)
		} else {
			valToPrint = fmt.Sprintf("(total %d)", fldData.selectedValsSum)
		}

		totalMarkOffset := x + leftOffset + lineLen + 1
		freeSpaceRight = width - (totalMarkOffset + len(valToPrint))
		if freeSpaceRight < 0 {
			totalMarkOffset += freeSpaceRight
		}
		tview.Print(screen, valToPrint, totalMarkOffset, y, width-totalMarkOffset, tview.AlignLeft, tcell.ColorLightGreen)
	}

	if h.externalCursorVisible {
		extCursorCoord := h.valToCoord(h.externalCursor)
		extCursorOffset := extCursorCoord / 2
		extCursorPos := x + fldMarginLeft + extCursorOffset
		extCursorRune, _ := ruleBuffer.Rune(extCursorPos)
		extCursorColor := "white"
		if extCursorRune == ' ' {
			extCursorRune = '^'
			extCursorColor = "white"
		}

		tview.Print(screen, fmt.Sprintf("[%s:red]%s[-:-]", extCursorColor, string(extCursorRune)), extCursorPos, y+height-1, width, tview.AlignLeft, tcell.ColorRed)
	}
}

func (h *Histogram) getDataBinsInChartBar() int {
	if h.fldData == nil {
		return 1
	}

	return h.fldData.dataBinsInChartBar
}

func (h *Histogram) getChartBarWidth() int {
	if h.fldData == nil {
		return 1
	}

	return h.fldData.chartBarWidth
}

func (h *Histogram) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return h.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		maxCursor := h.alignCursor(h.to-h.binSize*h.getDataBinsInChartBar(), true)

		moveLeft := func() {
			h.cursor = h.binSize * h.getDataBinsInChartBar()
			if h.cursor < h.from {
				h.cursor = h.from
			}
		}

		moveRight := func() {
			h.cursor += h.binSize * h.getDataBinsInChartBar()
			if h.cursor > maxCursor {
				h.cursor = maxCursor
			}
		}

		moveLeftLong := func() {
			target := h.from
			for _, mark := range h.curMarks {
				if mark >= h.cursor {
					break
				}
				target = mark
			}
			h.cursor = h.alignCursor(target, false)
		}

		moveRightLong := func() {
			moveRight()

			for _, mark := range h.curMarks {
				h.cursor = h.alignCursor(mark, false)
				return
			}
		}

		moveBegining := func() {
			h.cursor = h.from
		}

		moveEnd := func() {
			h.cursor = maxCursor
		}

		selectionEnd := func() {
			h.selectionStart = 0
		}

		selectionToggle := func() {
			if h.selectionStart == 0 {
				h.selectionStart = h.cursor
			} else {
				selectionEnd()
			}
		}

		selectionApplyIfActive := func() {
			if h.selectionStart != 0 && h.selected != nil {
				from, to := h.GetSelection()
				h.selected(from, to)
			}
		}

		switch event.Key() {
		case tcell.KeyRune:
			if event.Modifiers()&tcell.ModAlt > 0 {
				switch event.Rune() {
				case 'b':
					moveLeftLong()
				case 'f':
					moveRightLong()
				}
			} else {
				switch event.Rune() {
				case 'h':
					moveLeft()
				case 'l':
					moveRight()
				case 'b':
					moveLeftLong()
				case 'w', 'e':
					moveRightLong()
				case 'g', '^':
					moveBegining()
				case 'G', '$':
					moveEnd()
				case 'v', ' ':
					selectionApplyIfActive()
					selectionToggle()
				case 'q':
					selectionEnd()
				case 'o':
					if h.selectionStart > 0 {
						h.cursor, h.selectionStart = h.selectionStart, h.cursor
					}
				}
			}
		case tcell.KeyLeft:
			if event.Modifiers()&tcell.ModCtrl > 0 {
				moveLeftLong()
			} else {
				moveLeft()
			}
		case tcell.KeyRight:
			if event.Modifiers()&tcell.ModCtrl > 0 {
				moveRightLong()
			} else {
				moveRight()
			}
		case tcell.KeyPgUp:
			moveLeftLong()
		case tcell.KeyPgDn:
			moveRightLong()
		case tcell.KeyHome, tcell.KeyCtrlA:
			moveBegining()
		case tcell.KeyEnd, tcell.KeyCtrlE:
			moveEnd()
		case tcell.KeyEnter:
			selectionApplyIfActive()
			selectionToggle()
		case tcell.KeyEsc:
			selectionEnd()
		}
	})
}

func (h *Histogram) GetSelection() (selStart, selEnd int) {
	if h.selectionStart == 0 {
		return 0, 0
	}

	selStart = h.selectionStart
	selEnd = h.cursor
	if selStart > selEnd {
		selStart, selEnd = selEnd, selStart
	}

	selEnd += h.binSize * h.getDataBinsInChartBar()
	return selStart, selEnd
}

func (h *Histogram) SetSelectedFunc(handler func(from, to int)) *Histogram {
	h.selected = handler
	return h
}

func (h *Histogram) IsSelectionActive() bool {
	return h.selectionStart != 0
}

func (h *Histogram) getFieldData(width, height int) *fieldData {
	foc := h.HasFocus()
	scale := getOptimalScale(h.from, h.to, h.binSize, width, h.snapDataBinsInChartDot)
	if scale == nil {
		return nil
	}

	h.from = scale.from
	h.to = scale.to
	numDataBins := scale.numDataBins
	dataBinsInChartBar := scale.dataBinsInChartBar
	chartBarWidth := scale.chartBarWidth

	valAt := func(idx, n int) int {
		var val int
		for i := 0; i < n; i++ {
			val += h.data[int64(h.from+(idx+i)*h.binSize)]
		}
		return val
	}

	isCursorAt := func(idx, n int) bool {
		for i := 0; i < n; i++ {
			if h.cursor == h.from+(idx+i)*h.binSize {
				return true
			}
		}

		return false
	}

	selStart, selEnd := h.GetSelection()
	if selStart == 0 || selEnd == 0 {
		selStart = h.cursor
		selEnd = h.cursor + dataBinsInChartBar*h.binSize
	}

	isSelectedAt := func(idx, n int) bool {
		for i := 0; i < n; i++ {
			v := h.from + (idx+i)*h.binSize
			if v >= selStart && v < selEnd {
				return true
			}
		}
		return false
	}

	max := 0
	for xData := 0; xData < numDataBins; xData = xData + dataBinsInChartBar {
		val := valAt(xData, dataBinsInChartBar)
		if max < val {
			max = val
		}
	}

	dotYScale := (max + height - 1) / height
	dots := make([][]bool, height)
	for y := 0; y < height; y++ {
		dots[y] = make([]bool, width)
	}

	selScaleDots := make([][]bool, 2)
	for y := 0; y < 2; y++ {
		selScaleDots[y] = make([]bool, width)
	}

	selOffsetStart := -1
	selOffsetEnd := -1
	offsetLast := -1

	cursorVal := 0
	selectedValsSum := 0

	for xData, xChart := 0, 0; xData < numDataBins; xData, xChart = xData+dataBinsInChartBar, xChart+chartBarWidth {
		val := valAt(xData, dataBinsInChartBar)
		sel := isSelectedAt(xData, dataBinsInChartBar)
		crs := isCursorAt(xData, dataBinsInChartBar)

		if crs {
			cursorVal = val
		}
		if sel {
			selectedValsSum += val
		}

		for y := 0; y < height; y++ {
			on := val > y*dotYScale

			if !on && !(foc && sel) {
				break
			}

			if foc && sel {
				on = !on
			}

			if on {
				for i := 0; i < chartBarWidth; i++ {
					dots[height-y][xChart+i] = true
				}
			}
		}

		for i := 0; i < chartBarWidth; i++ {
			offsetLast = (xChart + i)
			if offsetLast&0x01 != 0 {
				offsetLast += 1
			}

			if sel {
				if selOffsetStart == -1 {
					selOffsetStart = (xChart + i)
					if selOffsetStart&0x01 != 0 {
						selOffsetStart = -1
					}
				}
				selScaleDots[0][xChart+i] = true
			} else if selOffsetStart != -1 && selOffsetEnd == -1 {
				selOffsetEnd = offsetLast
			}

			if crs {
				selScaleDots[1][xChart+i] = true
			}
		}
	}

	if selOffsetEnd == -1 {
		selOffsetEnd = offsetLast
	}
	for i := range selScaleDots {
		if selOffsetEnd != -1 {
			selScaleDots[i] = selScaleDots[i][:selOffsetEnd]
		}
		selScaleDots[i] = selScaleDots[i][selOffsetStart:]
	}

	effectiveWidthDots := numDataBins / dataBinsInChartBar * chartBarWidth
	effectiveWidthRunes := effectiveWidthDots / 2
	if effectiveWidthRunes&0x01 > 0 {
		effectiveWidthRunes++
	}

	return &fieldData{
		dots:                dots,
		dataBinsInChartBar:  dataBinsInChartBar,
		chartBarWidth:       chartBarWidth,
		max:                 max,
		dotYScale:           dotYScale,
		yScale:              dotYScale * height,
		effectiveWidthRunes: effectiveWidthRunes,
		effectiveWidthDots:  effectiveWidthDots,
		selScaleDots:        selScaleDots,
		selScaleOffset:      selOffsetStart,
		cursorVal:           cursorVal,
		selectedValsSum:     selectedValsSum,
	}
}

func getOptimalScale(from, to, binSize, width int, snapDataBinsInChartDot func(dataBinsInChartBar int) int) *histogramScale {
	if width < 0 {
		return nil
	}

	numDataBins := (to - from) / binSize
	if numDataBins == 0 {
		return nil
	}

	dataBinsInChartBar := (numDataBins + width - 1) / width
	dataBinsInChartBar = snapDataBinsInChartDot(dataBinsInChartBar)

	divisor := dataBinsInChartBar * binSize

	fromRemainder := from % divisor
	if fromRemainder > 0 {
		from -= fromRemainder
	}

	toRemainder := to % divisor
	if toRemainder > 0 {
		to += (divisor - toRemainder)
	}

	if fromRemainder > 0 || toRemainder > 0 {
		numDataBins = (to - from) / binSize
		dataBinsInChartBar = (numDataBins + width - 1) / width
		dataBinsInChartBar = snapDataBinsInChartDot(dataBinsInChartBar)
	}

	numBars := numDataBins / dataBinsInChartBar
	if numBars == 0 {
		return nil
	}

	chartBarWidth := width / numBars

	return &histogramScale{
		from:               from,
		to:                 to,
		numDataBins:        numDataBins,
		dataBinsInChartBar: dataBinsInChartBar,
		chartBarWidth:      chartBarWidth,
	}
}

func (h *Histogram) fldDataToLines(dots [][]bool) []string {
	ret := make([]string, 0, len(dots)/2)

	for y := 0; y < len(dots); y += 2 {
		fldRow1 := dots[y+0]
		fldRow2 := dots[y+1]

		row := strings.Builder{}
		row.Grow(len(fldRow1))

		for x := 0; x < len(fldRow1); x += 2 {
			qblockID := 0
			if fldRow1[x+0] {
				qblockID |= (1 << 3)
			}
			if fldRow1[x+1] {
				qblockID |= (1 << 2)
			}
			if fldRow2[x+0] {
				qblockID |= (1 << 1)
			}
			if fldRow2[x+1] {
				qblockID |= (1 << 0)
			}

			row.WriteRune(qblocks[qblockID])
		}

		ret = append(ret, row.String())
	}

	return ret
}

func (h *Histogram) valToCoord(v int) int {
	return (v - h.from) / h.getDataBinsInChartBar() * h.getChartBarWidth() / h.binSize
}

func clearViewFormatting(input string) string {
	var output strings.Builder
	inTag := false
	escaped := false
	bracketDepth := 0

	for i := 0; i < len(input); i++ {
		c := input[i]
		if inTag {
			if c == ']' {
				if escaped {
					output.WriteString("[" + input[bracketDepth:i-1] + "]")
					inTag = false
					escaped = false
				} else {
					inTag = false
				}
			} else if c == '[' && i+1 < len(input) && input[i+1] == ']' {
				escaped = true
			}
			continue
		}

		if c == '[' {
			if i+1 < len(input) && input[i+1] == '[' {
				output.WriteByte('[')
				i++
				continue
			}

			inTag = true
			bracketDepth = i + 1
			continue
		}
		output.WriteByte(c)
	}

	return output.String()
}

func highlightRune(s string, index int, prefix, suffix string) string {
	runes := []rune(s)
	if index >= 0 && index < len(runes) {
		return string(runes[:index]) + prefix + string(runes[index]) + suffix + string(runes[index+1:])
	}
	return s
}
