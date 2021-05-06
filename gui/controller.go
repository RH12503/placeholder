package main

import (
	"github.com/RH12503/Triangula/algorithm"
	"github.com/RH12503/Triangula/algorithm/evaluator"
	"github.com/RH12503/Triangula/generator"
	imageData "github.com/RH12503/Triangula/image"
	"github.com/RH12503/Triangula/mutation"
	"github.com/RH12503/Triangula/normgeom"
	"github.com/RH12503/tip-backend/save"
	"github.com/disintegration/imaging"
	"github.com/pterm/pterm"
	"github.com/wailsapp/wails"
	"image"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Controller struct {
	r     *wails.Runtime

	points, maxTime, maxSize int

	paths chan string

	nextId int
}

func (c *Controller) WailsInit(runtime *wails.Runtime) error {
	c.r = runtime

	c.paths = make(chan string)

	go func() {
		for path := range c.paths {
			c.r.Events.Emit("newPath", filepath.Base(path), c.nextId)
			c.nextId++
		}
	}()

	return nil
}

func (c *Controller) FilePressed() {
	path := c.r.Dialog.SelectFile("Select an image", "*.jpg,*.png,*.jpeg")

	if path == "" {
		return
	}

	c.addPath(path)
}

func (c *Controller) UpdatePressed(points, maxTime, maxSize int) {
	c.points = points
	c.maxTime = maxTime
	c.maxSize = maxSize
}

func (c *Controller) addPath(path string) {
	c.paths <- path
}



func processImage(imagePath string, numPoints int, timePerImage float64, maxSize int) error {

	file, err := os.Open(imagePath)

	if err != nil {
		return err
	}

	imageFile, _, err := image.Decode(file)
	file.Close()

	resizedImage := imageFile

	if maxSize != 0 {
		dim := imageFile.Bounds().Max
		if dim.X > dim.Y && dim.X > maxSize {
			resizedImage = imaging.Resize(imageFile, maxSize, 0, imaging.Lanczos)
		} else if dim.Y > dim.X && dim.Y > maxSize {
			resizedImage = imaging.Resize(imageFile, 0, maxSize, imaging.Lanczos)
		}
	}

	if err != nil {
		return err
	}
	img := imageData.ToData(resizedImage)

	pointFactory := func() normgeom.NormPointGroup {
		return generator.RandomGenerator{}.Generate(numPoints)
	}

	evaluatorFactory := func(n int) evaluator.Evaluator {
		return evaluator.NewParallel(img, 22, 5, n)
	}

	var mutator mutation.Method

	mutator = mutation.NewGaussianMethod(2./float64(numPoints), 0.3)

	algo := algorithm.NewModifiedGenetic(pointFactory, 400, 5, evaluatorFactory, mutator)

	ti := time.Now()

	for time.Since(ti).Seconds() < timePerImage {
		algo.Step()
	}

	ext := filepath.Ext(imagePath)

	filename := strings.TrimSuffix(filepath.Base(imagePath), ext) + ".tri"

	name := strings.TrimSuffix(imagePath, ext)

	save.WriteFile()
	if err := save.WriteFile(name+".tri", algo.Best(), imageData.ToData(imageFile)); err != nil {
		pterm.Error.WithShowLineNumber(false).Printf("Cannot write %v\n", filename)
		return err
	}
	return nil
}

func isDirectory(path string) bool {
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return true
	}
	return false
}

func validPath(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}
