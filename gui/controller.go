package main

import (
	"github.com/RH12503/Triangula/algorithm"
	"github.com/RH12503/Triangula/algorithm/evaluator"
	"github.com/RH12503/Triangula/fitness"
	"github.com/RH12503/Triangula/generator"
	imageData "github.com/RH12503/Triangula/image"
	"github.com/RH12503/Triangula/mutation"
	"github.com/RH12503/Triangula/normgeom"
	"github.com/RH12503/Triangula/render"
	"github.com/RH12503/Triangula/triangulation"
	"github.com/RH12503/tip-backend/save"
	"github.com/disintegration/imaging"
	"github.com/wailsapp/wails"
	"image"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Controller struct {
	r *wails.Runtime

	points, maxTime, maxSize int

	paths chan itemData

	nextId int

	currentId atomic.Value

	stopCurrent atomic.Value

	stopRun atomic.Value

	running atomic.Value

	remove      []int
	removeMutex sync.Mutex
}

func (c *Controller) WailsInit(runtime *wails.Runtime) error {

	c.r = runtime

	c.paths = make(chan itemData, 1024)

	c.running.Store(false)
	c.currentId.Store(-1)

	return nil
}

func (c *Controller) RemoveItem(id int) {
	if id == c.currentId.Load().(int) && c.running.Load().(bool) {
		c.stopCurrent.Store(true)
	} else {
		if id > c.currentId.Load().(int) {
			c.removeMutex.Lock()
			c.remove = append(c.remove, id)
			c.removeMutex.Unlock()
		}

		c.r.Events.Emit("remove", id)
	}
}

func (c *Controller) FilePressed() {
	path := c.r.Dialog.SelectFile("Select an image", "*.jpg,*.png,*.jpeg")

	if path == "" {
		return
	}

	c.addPath(path)
}

func (c *Controller) FolderPressed() {
	path := c.r.Dialog.SelectDirectory()

	if path == "" {
		return
	}

	c.addPath(path)
}

func (c *Controller) StartPressed(points, maxTime, maxSize int) {
	if c.running.Load().(bool) {
		c.stopRun.Store(true)
		c.stopCurrent.Store(true)
	} else {
		c.points = points
		c.maxTime = maxTime
		c.maxSize = maxSize
		if len(c.paths) != 0 {
			c.running.Store(true)
			c.stopRun.Store(false)
			go func() {
				c.r.Events.Emit("running")
				for info := range c.paths {
					c.removeMutex.Lock()
					if len(c.remove) != 0 && c.remove[0] == info.id {
						c.remove = c.remove[1:]
						c.removeMutex.Unlock()
						continue
					}
					c.removeMutex.Unlock()
					c.currentId.Store(info.id)
					func() {
						c.r.Events.Emit("working", info.id)
						file, err := os.Open(info.path)

						if err != nil {
							c.r.Events.Emit("error", info.id)
							return
						}

						imageFile, _, err := image.Decode(file)
						file.Close()

						resizedImage := imageFile

						if c.maxSize != 0 {
							dim := imageFile.Bounds().Max
							if dim.X > dim.Y && dim.X > c.maxSize {
								resizedImage = imaging.Resize(imageFile, c.maxSize, 0, imaging.Lanczos)
							} else if dim.Y > dim.X && dim.Y > c.maxSize {
								resizedImage = imaging.Resize(imageFile, 0, c.maxSize, imaging.Lanczos)
							}
						}

						if err != nil {
							c.r.Events.Emit("error", info.id)
							return
						}
						img := imageData.ToData(resizedImage)

						pointFactory := func() normgeom.NormPointGroup {
							return generator.RandomGenerator{}.Generate(c.points)
						}

						evaluatorFactory := func(n int) evaluator.Evaluator {
							return evaluator.NewParallel(fitness.TrianglesImageFunctions(img, 5, n), 22)
						}

						var mutator mutation.Method

						mutator = mutation.NewGaussianMethod(2./float64(c.points), 0.3)

						algo := algorithm.NewModifiedGenetic(pointFactory, 400, 5, evaluatorFactory, mutator)

						w, h := img.Size()

						c.stopCurrent.Store(false)
						ti := time.Now()
						d := 0.
						next := 0.
						for !c.stopCurrent.Load().(bool) && (c.maxTime <= 0 || d < float64(c.maxTime)) {
							if d > next {
								triangles := triangulation.Triangulate(algo.Best(), w, h)
								triangleData := render.TrianglesOnImage(triangles, img)

								c.r.Events.Emit("render", RenderData{
									Width:  w,
									Height: h,
									Data:   triangleData,
								})
								next = d + 1
							}

							algo.Step()
							d = time.Since(ti).Seconds()
							c.r.Events.Emit("time", info.id, d)
						}

						ext := filepath.Ext(info.path)

						name := strings.TrimSuffix(info.path, ext)

						if err := save.WriteFile(name+".tri", algo.Best(), imageData.ToData(imageFile)); err != nil {
							c.r.Events.Emit("error", info.id)
							return
						}
						c.r.Events.Emit("done", info.id)
					}()

					if c.stopRun.Load().(bool) || len(c.paths) == 0 {
						c.running.Store(false)
						break
					}
				}
				c.r.Events.Emit("stopped")
			}()
		}
	}
}

func (c *Controller) addPath(path string) {
	if !validPath(path) {
		return
	}

	if isDirectory(path) {
		files, err := ioutil.ReadDir(path)

		if err != nil {
			return
		}
		for _, f := range files {
			ext := filepath.Ext(f.Name())
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
				c.r.Events.Emit("newPath", f.Name(), c.nextId)
				c.paths <- itemData{
					path: filepath.Join(path, f.Name()),
					id:   c.nextId,
				}
				c.nextId++
			}
		}
	} else {
		c.r.Events.Emit("newPath", filepath.Base(path), c.nextId)
		c.paths <- itemData{
			path: path,
			id:   c.nextId,
		}
		c.nextId++
	}
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

type RenderData struct {
	Width, Height int
	Data          []render.TriangleData
}

type itemData struct {
	path string
	id   int
}
