package shell

import (
	"strings"
	"unicode"

	"nte-tools/internal/ui/common"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// messageText 是状态栏里的一段消息：按自己的实际宽度折行显示，行数变化时更新最小
// 高度并请父级重新布局，所以状态栏会随行数变高，而不是把消息裁掉。
// Fyne 的标签不会把折行算进最小尺寸，因此这里先自己折行，再把折好的文字交给标签画。
type messageText struct {
	widget.BaseWidget

	text    string
	wrapped string
	lines   int
}

// newMessageText 构建一段空消息。
func newMessageText() *messageText {
	message := &messageText{lines: 1}
	message.ExtendBaseWidget(message)

	return message
}

// setText 换一段消息，先按当前宽度折行。
func (m *messageText) setText(text string) {
	m.text = text
	m.rewrap(m.Size().Width)

	m.Refresh()
}

// MinSize 返回所需尺寸：宽度由父级决定，高度按折行后的行数。
func (m *messageText) MinSize() fyne.Size {
	m.ExtendBaseWidget(m)

	lines := m.lines
	if lines < 1 {
		lines = 1
	}

	return fyne.NewSize(0, float32(lines)*messageLineHeight())
}

// rewrap 按宽度重新折行；折行结果变了就更新最小高度并请父级重新布局。
func (m *messageText) rewrap(width float32) {
	wrapped, lines := wrapMessage(m.text, width)

	if wrapped == m.wrapped && lines == m.lines {
		return
	}

	m.wrapped = wrapped
	m.lines = lines

	canvas.Refresh(m)
}

// CreateRenderer 创建消息的绘制对象：一个显示折好行文字的小号标签。
func (m *messageText) CreateRenderer() fyne.WidgetRenderer {
	label := widget.NewLabel("")
	label.Importance = widget.LowImportance
	label.SizeName = theme.SizeNameCaptionText

	renderer := &messageTextRenderer{
		BaseRenderer: common.NewBaseRenderer(label),
		message:      m,
		label:        label,
	}
	renderer.Refresh()

	return renderer
}

type messageTextRenderer struct {
	common.BaseRenderer

	message *messageText
	label   *widget.Label
}

func (r *messageTextRenderer) Refresh() {
	r.label.SetText(r.message.wrapped)

	canvas.Refresh(r.message)
}

func (r *messageTextRenderer) Layout(size fyne.Size) {
	r.message.rewrap(size.Width)

	r.label.Move(fyne.NewPos(0, 0))
	r.label.Resize(size)
}

func (r *messageTextRenderer) MinSize() fyne.Size {
	return r.message.MinSize()
}

// messageTextSize 返回消息用的字号。
func messageTextSize() float32 {
	th, _ := themeColors()

	return th.Size(theme.SizeNameCaptionText)
}

// messageLineHeight 返回消息单行的高度。
func messageLineHeight() float32 {
	return fyne.MeasureText("汉Ag", messageTextSize(), fyne.TextStyle{}).Height
}

// wrapMessage 按可用宽度把文本折成多行，返回折好的文本与行数。宽度未知时按一行算；
// 优先在空白处断开，找不到就按字符断开。
func wrapMessage(text string, width float32) (string, int) {
	if text == "" {
		return "", 1
	}

	if width <= 0 {
		return text, 1
	}

	textSize := messageTextSize()
	style := fyne.TextStyle{}
	runes := []rune(text)

	var builder strings.Builder

	lines := 0

	for len(runes) > 0 {
		fit := fitRunes(runes, width, textSize, style)

		if cut := lastSpaceIndex(runes[:fit]); cut > 0 {
			fit = cut
		}

		line := strings.TrimSpace(string(runes[:fit]))

		if line != "" {
			if lines > 0 {
				builder.WriteByte('\n')
			}

			builder.WriteString(line)
			lines++
		}

		runes = runes[fit:]
	}

	if lines == 0 {
		return "", 1
	}

	return builder.String(), lines
}

// fitRunes 用二分找出前多少个字符能放进 width（至少一个）。
func fitRunes(runes []rune, width, textSize float32, style fyne.TextStyle) int {
	low, high := 1, len(runes)

	for low < high {
		middle := (low + high + 1) / 2

		if fyne.MeasureText(string(runes[:middle]), textSize, style).Width <= width {
			low = middle
		} else {
			high = middle - 1
		}
	}

	return low
}

// lastSpaceIndex 返回最后一段空白之后的下标；没有空白时返回 0。
func lastSpaceIndex(runes []rune) int {
	for i := len(runes) - 1; i > 0; i-- {
		if unicode.IsSpace(runes[i-1]) {
			return i
		}
	}

	return 0
}
