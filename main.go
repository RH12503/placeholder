package main

import (
	"fmt"
	"github.com/RH12503/Triangula/algorithm"
	"github.com/RH12503/Triangula/algorithm/evaluator"
	"github.com/RH12503/Triangula/generator"
	imageData "github.com/RH12503/Triangula/image"
	"github.com/RH12503/Triangula/mutation"
	"github.com/RH12503/Triangula/normgeom"
	"github.com/disintegration/imaging"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {

	app := &cli.App{
		Name:  "placeholder",
		Usage: "",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "input",
				Usage:    "An image or a directory with images to process.",
				Aliases:  []string{"i", "in"},
				Required: true,
			},
			&cli.Float64Flag{
				Name:     "time",
				Usage:    "How many seconds to run the algorithm for each image.",
				Aliases:  []string{"t"},
				Value: 60,
			},
			&cli.UintFlag{
				Name:     "points",
				Usage:    "How many points to use in each triangulation.",
				Aliases:  []string{"p", "pts"},
				Value: 600,
			},
			&cli.UintFlag{
				Name:    "max-size",
				Usage:   "The maximum size in pixels to clamp the images to.",
				Aliases: []string{"s", "ms"},
			},
		},
		Action: func(c *cli.Context) error {
			numPoints := int(c.Uint("points"))
			timePerImage := c.Float64("time")
			path := c.String("input")
			maxSize := int(c.Uint("max-size"))

			parameters := fmt.Sprintf("[Points] %v | [Time] %vs", numPoints, timePerImage)

			if maxSize != 0 {
				parameters += fmt.Sprintf(" | [Max size] %vpx", maxSize)
			}

			pterm.Info.Println(parameters)

			if !validPath(path) {
				pterm.Error.WithShowLineNumber(false).Println("Invalid path")
				return nil
			}

			if isDirectory(path) {
				files, err := ioutil.ReadDir(path)
				if err != nil {
					pterm.Error.WithShowLineNumber(false).Println("Cannot read directory")
				}

				var paths []string
				for _, f := range files {
					ext := filepath.Ext(f.Name())
					if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
						paths = append(paths, filepath.Join(path, f.Name()))
					}
				}

				if len(paths) == 0 {
					pterm.Warning.Println("No images found!")
					return nil
				}
				pterm.Info.Printf("%v images found!\n", len(paths))

				bar, _ := pterm.DefaultProgressbar.WithTotal(len(paths)).WithTitle("Processing images").Start()
				processed := 0
				for _, p := range paths {
					err := processImage(p, numPoints, timePerImage, maxSize)
					if err == nil {
						processed++
						pterm.Success.Printf("Processed %v!\n", filepath.Base(p))
					}
					bar.Increment()
				}
				if processed == len(paths) {
					pterm.Success.Println("All images processed!")
				} else {
					pterm.Warning.Printf("%v/%v images could not be processed!\n", len(paths)-processed, len(paths))
				}

			} else {
				pterm.Info.Printf("Processing %v\n", filepath.Base(path))
				err := processImage(path, numPoints, timePerImage, maxSize)
				if err == nil {
					pterm.Success.Printf("Processed %v!\n", filepath.Base(path))
				}
			}

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}

func processImage(imagePath string, numPoints int, timePerImage float64, maxSize int) error {

	file, err := os.Open(imagePath)

	if err != nil {
		pterm.Error.WithShowLineNumber(false).Printf("Cannot read %v\n", filepath.Base(imagePath))
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
		pterm.Error.WithShowLineNumber(false).Printf("Cannot decode %v\n", filepath.Base(imagePath))
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

	if err := SaveFile(name+".tri", algo.Best(), imageData.ToData(imageFile)); err != nil {
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
