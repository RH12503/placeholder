package save

import (
	"bufio"
	"encoding/binary"
	"github.com/RH12503/Triangula/color"
	"github.com/RH12503/Triangula/geom"
	"github.com/RH12503/Triangula/image"
	"github.com/RH12503/Triangula/normgeom"
	"github.com/RH12503/Triangula/render"
	"github.com/RH12503/Triangula/triangulation"
	"math"
	"os"
	"sort"
)

func WriteFile(filepath string, points normgeom.NormPointGroup, image image.Data) error {

	w, h := image.Size()

	triangles := triangulation.Triangulate(points, w, h)
	renderData := render.TrianglesOnImage(triangles, image)

	sortTriangles(triangles)

	var active []int
	visited := make(map[int]bool)
	faces := make(map[int]faceData)
	parent := make(map[int]int)

	file, err := os.Create(filepath)
	defer file.Close()

	if err != nil {
		return err
	}

	writer := bufio.NewWriter(file)

	binary.Write(writer, binary.LittleEndian, uint16(w))
	binary.Write(writer, binary.LittleEndian, uint16(h))

	first := 0

	writer.Write([]byte{uint8(numberAdjacent(triangles[first], triangles))})
	for _, p := range triangles[first].Points {
		binary.Write(writer, binary.LittleEndian, uint16(p.X))
		binary.Write(writer, binary.LittleEndian, uint16(p.Y))
	}
	col := renderData[first].Color
	writer.Write([]byte{uint8(multAndRound(col.R, 255))})
	writer.Write([]byte{uint8(multAndRound(col.G, 255))})
	writer.Write([]byte{uint8(multAndRound(col.B, 255))})

	active = append(active, first)
	visited[first] = true

	notFirst := false

	for len(active) > 0 {
		index := active[0]
		current := triangles[index]
		active = active[1:]

		n := 0
		for i, b := range triangles {
			if adj, a, b := adjacent(current, b); adj && !visited[i] {
				active = append(active, i)
				visited[i] = true
				faces[i] = faceData{a: a, b: b}
				parent[i] = index
				n++
			}
		}

		if notFirst {
			col := renderData[index].Color

			v := 0
			if faces[index].b == 2 {
				v = 1
			} else if faces[index].b == 0 {
				v = 2
			}

			parentPoint := triangles[parent[index]].Points[faces[index].a]
			diff := coordsDiff(parentPoint, current.Points[v])
			compressCoords := diff.X <= 127 && diff.X >= -128 && diff.Y <= 127 && diff.Y >= -128

			dR, dG, dB := colorDiff(renderData[parent[index]].Color, col)
			compressColor := dR <= 31 && dR >= -32 && dG <= 31 && dG >= -32 && dB <= 31 && dB >= -32

			r := uint16(dR + 32)
			g := uint16(dG + 32)
			b := uint16(dB + 32)
			dataByte := uint8(n) | uint8(faces[index].a)<<2 | boolToByte(compressCoords)<<4 | boolToByte(compressColor)<<5 | ((uint8(b)>>4)&3)<<6
			writer.Write([]byte{dataByte})

			if compressCoords {
				writer.Write([]byte{byte(diff.X), byte(diff.Y)})
			} else {
				p := current.Points[v]
				binary.Write(writer, binary.LittleEndian, uint16(p.X))
				binary.Write(writer, binary.LittleEndian, uint16(p.Y))
			}

			if compressColor {
				colorData := r | g<<6 | (b&15)<<12

				binary.Write(writer, binary.LittleEndian, colorData)
			} else {
				writer.Write([]byte{uint8(multAndRound(col.R, 255))})
				writer.Write([]byte{uint8(multAndRound(col.G, 255))})
				writer.Write([]byte{uint8(multAndRound(col.B, 255))})
			}
		}

		notFirst = true
	}

	writer.Flush()

	return nil
}

func sortTriangles(triangles []geom.Triangle) {
	for t := range triangles {
		sort.Slice(triangles[t].Points[:], func(i, j int) bool {
			if triangles[t].Points[i].Y == triangles[t].Points[j].Y {
				return triangles[t].Points[i].X < triangles[t].Points[j].X
			}

			return triangles[t].Points[i].Y < triangles[t].Points[j].Y
		})
	}
}

func numberAdjacent(triangle geom.Triangle, triangles []geom.Triangle) int {
	n := 0
	for _, b := range triangles {
		if adj, _, _ := adjacent(triangle, b); adj {
			n++
		}
	}
	return n
}

func adjacent(a, b geom.Triangle) (bool, int, int) {
	common := 0
	sumA := 0
	sumB := 0
	for i, pA := range a.Points {
		for j, pB := range b.Points {
			if pA == pB {
				sumA += i
				sumB += j
				common++
			}
		}
	}

	return common == 2, faceFromSum(sumA), faceFromSum(sumB)
}

func faceFromSum(sum int) int {
	var face int

	if sum == 1 {
		face = 0
	} else if sum == 3 {
		face = 1
	} else if sum == 2 {
		face = 2
	}
	return face
}
func boolToByte(b bool) uint8 {
	if b {
		return 1
	}

	return 0
}

func coordsDiff(a, b geom.Point) geom.Point {
	return geom.Point{
		X: b.X - a.X,
		Y: b.Y - a.Y,
	}
}

func colorDiff(a, b color.RGB) (int, int, int) {
	dR := multAndRound(b.R, 255) - multAndRound(a.R, 255)
	dG := multAndRound(b.G, 255) - multAndRound(a.G, 255)
	dB := multAndRound(b.B, 255) - multAndRound(a.B, 255)
	return dR, dG, dB
}

func multAndRound(v float64, m int) int {
	return int(math.Round(v * float64(m)))
}

type faceData struct {
	a, b int
}
