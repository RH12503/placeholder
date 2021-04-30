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

func SaveFile(filepath string, points normgeom.NormPointGroup, image image.Data) error {
	w, h := image.Size()

	triangles := triangulation.Triangulate(points, w, h)
	renderData := render.TrianglesOnImage(triangles, image)

	sortTriangles(triangles)

	var active []int
	visited := make(map[int]bool)
	faces := make(map[int]faceData)

	file, err := os.Create(filepath)
	defer file.Close()

	if err != nil {
		return err
	}

	writer := bufio.NewWriter(file)

	writer.Write(uint16ToBytes(uint16(w)))
	writer.Write(uint16ToBytes(uint16(h)))

	first := 0

	writer.Write([]byte{uint8(numberAdjacent(triangles[first], triangles))})
	for _, p := range triangles[first].Points {
		writer.Write(uint16ToBytes(uint16(p.X)))
		writer.Write(uint16ToBytes(uint16(p.Y)))
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
				n++
			}
		}

		if notFirst {
			col := renderData[index].Color
			writer.Write([]byte{uint8(3*n + faces[index].a)})

			v := 0
			if faces[index].b == 2 {
				v = 1
			} else if faces[index].b == 0 {
				v = 2
			}

			writer.Write(uint16ToBytes(uint16(current.Points[v].X)))
			writer.Write(uint16ToBytes(uint16(current.Points[v].Y)))

			writer.Write([]byte{uint8(multAndRound(col.R, 255))})
			writer.Write([]byte{uint8(multAndRound(col.G, 255))})
			writer.Write([]byte{uint8(multAndRound(col.B, 255))})
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

func uint16ToBytes(num uint16) []byte {
	bytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(bytes, num)
	return bytes
}

func multAndRound(v float64, m int) int {
	return int(math.Round(v * float64(m)))
}

type faceData struct {
	a, b int
}
