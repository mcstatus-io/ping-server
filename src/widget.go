package main

import (
	"fmt"
	"image/color"

	"github.com/fogleman/gg"
)

const (
	// WidgetWidth is the width of the image of the widget.
	WidgetWidth = 860
	// WidgetHeight is the height of the image of the widget.
	WidgetHeight = 240
)

// GenerateJavaWidget generates a widget image from the Java Edition status response.
func GenerateJavaWidget(status *JavaStatusResponse, isDark bool) ([]byte, error) {
	var statusColor color.Color = color.RGBA{230, 57, 70, 255}
	var statusText string = "Offline"
	var players string = "Unknown players"
	var address string = status.Host

	if status.Online {
		statusColor = color.RGBA{46, 204, 113, 255}
		statusText = "Online"

		if status.Players.Online != nil && status.Players.Max != nil {
			players = fmt.Sprintf("%d/%d players", *status.Players.Online, *status.Players.Max)
		} else if status.Players.Online != nil {
			players = fmt.Sprintf("%d players", *status.Players.Online)
		} else {
			players = "Unknown players"
		}
	}

	if status.Port != 25565 {
		address += fmt.Sprintf(":%d", status.Port)
	}

	icon, err := GetStatusIcon(status)

	if err != nil {
		return nil, err
	}

	ctx := gg.NewContext(WidgetWidth, WidgetHeight)

	// Background color
	if isDark {
		ctx.SetColor(color.Gray{32})
	} else {
		ctx.SetColor(color.White)
	}
	ctx.DrawRoundedRectangle(0, 0, WidgetWidth, WidgetHeight, 16)
	ctx.Fill()

	// Draw server icon
	ctx.DrawImage(ScaleImageNearestNeighbor(icon, 3, 3), (WidgetHeight-64*3)/2, (WidgetHeight-64*3)/2)

	// Draw status bubble
	ctx.SetColor(statusColor)
	ctx.DrawCircle(WidgetHeight+8, WidgetHeight/2-55, 8)
	ctx.Fill()

	// Draw status text
	ctx.SetFontFace(ubuntuBoldFont)
	ctx.DrawString(statusText, WidgetHeight+28, WidgetHeight/2-44)
	ctx.Stroke()

	// Draw address text
	if isDark {
		ctx.SetColor(color.White)
	} else {
		ctx.SetColor(color.Black)
	}

	ctx.SetFontFace(ubuntuMonoFont)
	ctx.DrawString(address, WidgetHeight, WidgetHeight/2+4)
	ctx.Stroke()

	// Draw players text
	ctx.SetColor(color.Gray{127})
	ctx.SetFontFace(ubuntuRegularFont)
	ctx.DrawString(players, WidgetHeight, WidgetHeight/2+58)
	ctx.Stroke()

	// Draw mcstatus.io branding
	ctx.SetColor(color.Gray{152})
	ctx.SetFontFace(ubuntuMonoSmallFont)
	ctx.DrawStringAnchored("© mcstatus.io", WidgetWidth-16, WidgetHeight-38, 1.0, 1.0)
	ctx.Stroke()

	return EncodePNG(ctx.Image())
}
