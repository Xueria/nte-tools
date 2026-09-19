package listing

import (
	"blind-tools/internal/bid"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

// previewCellSize 占格预览里单个格子的边长。
const previewCellSize float32 = 34

// listingPreview 返回一份清单在预览区的呈现内容：带占格属性的清单画出
// length×width 的格子，普通拍品清单没有预览可画。清单标题一律直接用清单自己的
// 名称，所以这里只决定预览内容；新增清单类型时在这里补一个分支即可，
// 页面自身不必知道有哪些类型。
func listingPreview(listing bid.Listing) fyne.CanvasObject {
	if listing.Attribute.Grid {
		return footprint(listing.Attribute.Length, listing.Attribute.Width)
	}

	return nil
}

// previewObjects 把可选的预览内容包成预览区的内容；没有预览时预览区为空。
func previewObjects(preview fyne.CanvasObject) []fyne.CanvasObject {
	if preview == nil {
		return nil
	}

	return []fyne.CanvasObject{preview}
}

// footprint 画出 length 列、width 行的占格形状。
func footprint(length, width int) fyne.CanvasObject {
	cells := make([]fyne.CanvasObject, 0, length*width)

	for i := 0; i < length*width; i++ {
		cells = append(cells, previewCell())
	}

	// 居中放置：格子保持正方，不被布局拉伸。
	return container.NewCenter(container.NewGridWithColumns(length, cells...))
}

// previewCell 生成占格预览里的一个格子。
func previewCell() fyne.CanvasObject {
	settings := fyne.CurrentApp().Settings()
	variant := settings.ThemeVariant()

	cell := canvas.NewRectangle(settings.Theme().Color(theme.ColorNameInputBackground, variant))
	cell.StrokeColor = settings.Theme().Color(theme.ColorNameInputBorder, variant)
	cell.StrokeWidth = 1
	cell.CornerRadius = 4
	cell.SetMinSize(fyne.NewSize(previewCellSize, previewCellSize))

	return cell
}
