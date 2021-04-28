package main

import (
	"bufio"
	"encoding/binary"
	"github.com/RH12503/Triangula/geom"
	"github.com/RH12503/Triangula/image"
	"github.com/RH12503/Triangula/normgeom"
	"github.com/RH12503/Triangula/render"
	"github.com/RH12503/Triangula/triangulation"
	"math"
	"os"
	"sort"
)

func SaveFile2(filepath string, points normgeom.NormPointGroup, image image.Data) error {
	w, h := image.Size()

	triangles := triangulation.Triangulate(points, w, h)
	distances := make([]int, len(triangles))
	//triangleData := render.TrianglesOnImage(triangles, image)

	for i := range distances {
		distances[i] = -1
	}

	layer := 0
	var active []int
	visited := make(map[geom.Triangle]bool)

	for i, t := range triangles {
		n := 0
		for _, b := range triangles {
			if adjacent(t, b) {
				n++
			}
		}

		if n != 3 {
			active = append(active, i)
			distances[i] = layer
			visited[t] = true
		}
	}

	var nextActive []int

	for len(visited) < len(triangles) {
		layer++

		for len(active) > 0 {
			current := active[0]
			active = active[1:]

			for i, t := range triangles {
				if adjacent(t, triangles[current]) {
					if !visited[t] {
						nextActive = append(nextActive, i)
						distances[i] = layer
						visited[t] = true
					}
				}
			}
		}
		active, nextActive = nextActive, active
	}

	tris := make([]struct {
		t    geom.Triangle
		dist int
	}, len(triangles))

	for i := range tris {
		tris[i] = struct {
			t    geom.Triangle
			dist int
		}{t: triangles[i], dist: distances[i]}
	}

	sort.Slice(tris, func(i, j int) bool {
		return tris[i].dist < tris[j].dist
	})

	visited = make(map[geom.Triangle]bool)

	current := tris[0]

	var path []struct {
		tri struct {
			t    geom.Triangle
			dist int
		}
		b int
	}

Outer:
	for len(visited) < len(triangles) {

		visited[current.t] = true

		for _, t := range tris {
			if adjacent(current.t, t.t) && !visited[t.t] {
				current = t
				path = append(path, struct {
					tri struct {
						t    geom.Triangle
						dist int
					}
					b int
				}{tri: current, b: len(path) - 1})
				continue Outer
			}
		}
		current = path[path[len(path)-1].b].tri
		path = append(path, struct {
			tri struct {
				t    geom.Triangle
				dist int
			}
			b int
		}{tri: current, b: path[len(path)-1].b})

	}

	file, err := os.Create(filepath)
	defer file.Close()

	if err != nil {
		return err
	}

	writer := bufio.NewWriter(file)

	writer.Write(uint16ToBytes(uint16(w)))
	writer.Write(uint16ToBytes(uint16(h)))

	/*for _, d := range triangleData {
		tri := d.Triangle.Points
		col := d.Color

		for _, p := range tri {
			point := geom.Point{
				X: multAndRound(p.X, w),
				Y: multAndRound(p.Y, h),
			}

			writer.Write(uint16ToBytes(uint16(pointsMap[point])))
		}

		writer.Write([]byte{uint8(multAndRound(col.R, 255))})
		writer.Write([]byte{uint8(multAndRound(col.G, 255))})
		writer.Write([]byte{uint8(multAndRound(col.B, 255))})
	}

	writer.Flush()*/

	return nil
}

func adjacent(a, b geom.Triangle) bool {
	common := 0

	for _, pA := range a.Points {
		for _, pB := range b.Points {
			if pA == pB {
				common++
			}
		}
	}

	return common == 2
}

func averageY(tri geom.Triangle) float64 {
	y := tri.Points[0].Y + tri.Points[1].Y + tri.Points[2].Y

	return float64(y) / 3
}

func SaveFile(filepath string, points normgeom.NormPointGroup, image image.Data) error {
	w, h := image.Size()

	triangles := triangulation.Triangulate(points, w, h)
	triangleData := render.TrianglesOnImage(triangles, image)

	file, err := os.Create(filepath)
	defer file.Close()

	if err != nil {
		return err
	}

	writer := bufio.NewWriter(file)

	writer.Write(uint16ToBytes(uint16(w)))
	writer.Write(uint16ToBytes(uint16(h)))

	writer.Write(uint16ToBytes(uint16(len(points))))

	pointsMap := make(map[geom.Point]int)

	for i, p := range points {
		point := geom.Point{
			X: multAndRound(p.X, w),
			Y: multAndRound(p.Y, h),
		}
		writer.Write(uint16ToBytes(uint16(point.X)))
		writer.Write(uint16ToBytes(uint16(point.Y)))

		pointsMap[point] = i
	}

	for _, d := range triangleData {
		tri := d.Triangle.Points
		col := d.Color

		for _, p := range tri {
			point := geom.Point{
				X: multAndRound(p.X, w),
				Y: multAndRound(p.Y, h),
			}

			writer.Write(uint16ToBytes(uint16(pointsMap[point])))
		}

		writer.Write([]byte{uint8(multAndRound(col.R, 255))})
		writer.Write([]byte{uint8(multAndRound(col.G, 255))})
		writer.Write([]byte{uint8(multAndRound(col.B, 255))})
	}

	writer.Flush()

	return nil
}

func uint16ToBytes(num uint16) []byte {
	bytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(bytes, num)
	return bytes
}

func multAndRound(v float64, m int) int {
	return int(math.Round(v * float64(m)))
}
