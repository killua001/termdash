package container

// draw.go contains logic to draw containers and the contained widgets.

import (
	"errors"
	"fmt"
	"image"

	"github.com/mum4k/termdash/area"
	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/draw"
)

// drawTree draws this container and all of its sub containers.
func drawTree(c *Container) error {
	var errStr string
	preOrder(c, &errStr, visitFunc(func(c *Container) error {
		return drawCont(c)
	}))
	if errStr != "" {
		return errors.New(errStr)
	}
	return nil
}

// drawBorder draws the border around the container if requested.
func drawBorder(c *Container) error {
	if !c.hasBorder() {
		return nil
	}

	cvs, err := canvas.New(c.area)
	if err != nil {
		return err
	}

	ar, err := area.FromSize(cvs.Size())
	if err != nil {
		return err
	}

	var opts []cell.Option
	if c.focusTracker.isActive(c) {
		opts = append(opts, cell.FgColor(c.opts.inherited.focusedColor))
	} else {
		opts = append(opts, cell.FgColor(c.opts.inherited.borderColor))
	}
	if err := draw.Box(cvs, ar, c.opts.border, opts...); err != nil {
		return err
	}
	return cvs.Apply(c.term)
}

// hAlignWidget adjusts the given widget area within the containers area
// based on the requested horizontal alignment.
func hAlignWidget(c *Container, wArea image.Rectangle) image.Rectangle {
	gap := c.usable().Dx() - wArea.Dx()
	switch c.opts.hAlign {
	case hAlignTypeRight:
		// Use gap from above.
	case hAlignTypeCenter:
		gap /= 2
	default:
		// Left or unknown.
		gap = 0
	}

	return image.Rect(
		wArea.Min.X+gap,
		wArea.Min.Y,
		wArea.Max.X+gap,
		wArea.Max.Y,
	)
}

// vAlignWidget adjusts the given widget area within the containers area
// based on the requested vertical alignment.
func vAlignWidget(c *Container, wArea image.Rectangle) image.Rectangle {
	gap := c.usable().Dy() - wArea.Dy()
	switch c.opts.vAlign {
	case vAlignTypeBottom:
		// Use gap from above.
	case vAlignTypeMiddle:
		gap /= 2
	default:
		// Top or unknown.
		gap = 0
	}

	return image.Rect(
		wArea.Min.X,
		wArea.Min.Y+gap,
		wArea.Max.X,
		wArea.Max.Y+gap,
	)
}

// drawWidget requests the widget to draw on the canvas.
func drawWidget(c *Container) error {
	widgetArea := c.widgetArea()
	if widgetArea == image.ZR {
		return nil
	}

	if !c.hasWidget() {
		return nil
	}

	needSize := image.Point{1, 1}
	wOpts := c.opts.widget.Options()
	if wOpts.MinimumSize.X > 0 && wOpts.MinimumSize.Y > 0 {
		needSize = wOpts.MinimumSize
	}

	if widgetArea.Dx() < needSize.X || widgetArea.Dy() < needSize.Y {
		return drawResize(c, c.usable())
	}

	cvs, err := canvas.New(widgetArea)
	if err != nil {
		return err
	}

	if err := c.opts.widget.Draw(cvs); err != nil {
		return err
	}
	return cvs.Apply(c.term)
}

// drawResize draws an unicode character indicating that the size is too small to draw this container.
// Does nothing if the size is smaller than one cell, leaving no space for the character.
func drawResize(c *Container, area image.Rectangle) error {
	if area.Dx() < 1 || area.Dy() < 1 {
		return nil
	}

	cvs, err := canvas.New(area)
	if err != nil {
		return err
	}

	if err := draw.Text(cvs, "⇄", draw.TextBounds{}); err != nil {
		return err
	}
	return cvs.Apply(c.term)
}

// drawCont draws the container and its widget.
func drawCont(c *Container) error {
	if us := c.usable(); us.Dx() <= 0 || us.Dy() <= 0 {
		return drawResize(c, c.area)
	}

	if err := drawBorder(c); err != nil {
		return fmt.Errorf("unable to draw container border: %v", err)
	}

	if err := drawWidget(c); err != nil {
		return fmt.Errorf("unable to draw widget %T: %v", c.opts.widget, err)
	}
	return nil
}