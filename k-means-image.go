////////////////////////////////////////
//Colt Darien                         //
//nearestNeighbor.go                  //
////////////////////////////////////////
//This program uses a modified k-means clustering to find information about what colors exist in an image.

package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"math"
	"math/rand"
	"os"
	"strconv"
)

////////////////////////////////////////
//Cluster//
////////////////////////////////////////

//cluster contains the information for center of the cluster, etc
type cluster struct {
	X, Y, Z     int64 //X, Y, Z are basically R, G, B values of all members averaged together
	Members     int64 //the number of points that are members of this cluster
	colorSpread int64 //used in the color brightening step
}

func (c *cluster) String() string {
	if c == nil {
		return "nil"
	}
	//image outputs RGB values in 32 bits (16 for the color, and another 16 for overflow from math), we need to get them to be in the 0-255 scale again)
	x := int(float64(c.X) / 65535 * 255)
	y := int(float64(c.Y) / 65535 * 255)
	z := int(float64(c.Z) / 65535 * 255)

	var str string = "{(" + strconv.Itoa(x) + ", " + strconv.Itoa(y) + ", " + strconv.Itoa(z) + ") Members: " + strconv.Itoa(int(c.Members)) + "}"
	return str
}

////////////////////////////////////////
//Point//
////////////////////////////////////////
//point will contain the information for each point (colors) and
type point struct {
	X      uint32
	Y      uint32
	Z      uint32
	spread int
	Center *cluster
}

func (p *point) getDist(cluster *cluster) int64 {
	xDist := (cluster.X - int64(p.X))
	yDist := (cluster.Y - int64(p.Y))
	zDist := (cluster.Z - int64(p.Z))
	//suareDist is as good as actual distance for comparrisons, no need for sqrt
	squaredDist := (xDist * xDist) + (yDist * yDist) + (zDist * zDist)
	//squaredDist := xDist + yDist + zDist
	return squaredDist
}

func (p *point) findClosest(clusters []*cluster) *cluster {
	//initialie the closest
	var closestCluster *cluster = clusters[0]
	var minSquaredDist int64

	minSquaredDist = p.getDist(closestCluster)

	for _, cluster := range clusters[1:] {
		squaredDist := p.getDist(cluster)
		if squaredDist < minSquaredDist {
			minSquaredDist = squaredDist
			closestCluster = cluster
		}
	}
	return closestCluster
}

//assignCenter will take a point as an argument, and add it to the center. assignCenter returns the distance that the center moved. TODO is int working?
func (p *point) assignCenter(c *cluster) int {
	//update members
	c.Members++
	//add in the new values
	c.X += (int64(p.X) - c.X) / c.Members
	c.Y += (int64(p.Y) - c.Y) / c.Members
	c.Z += (int64(p.Z) - c.Z) / c.Members
	//update the points center
	p.Center = c
	//TODO return distance that it moved as per optimizations
	return 0
}

//unassignCenter will take a point as an argument, and add it to the center. unassignCenter returns the distance that the center moved. TODO is int return working?
func (p *point) unassignCenter() int {
	if p.Center == nil {
		return 0
	}
	if p.Center.Members > 1 {
		//update members
		p.Center.Members--
		//get rid of the old value
		p.Center.X -= (int64(p.X) - p.Center.X) / p.Center.Members
		p.Center.Y -= (int64(p.Y) - p.Center.Y) / p.Center.Members
		p.Center.Z -= (int64(p.Z) - p.Center.Z) / p.Center.Members
	} else {
		//we cannot adjust the location any lower
		p.Center.Members--
	}
	//update the points p.Center
	p.Center = nil
	//TODO return distance that it moved as per optimizations
	return 0
}

func getPointsFromImage(imageName string) []*point {
	reader, err := os.Open(imageName)
	check(err)
	m, _, err := image.Decode(reader)
	check(err)
	bounds := m.Bounds()
	var points []*point
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := m.At(x, y).RGBA()
			var s int = 0
			//TODO I average the two non max colors, maybe I should try max - min instead.
			if r > g && r > b {
				s = int(r) - int((g+b)/2)
			}
			if g > r && g > b {
				s = int(g) - int((r+b)/2)
			}
			if b > r && b > g {
				s = int(b) - int((g+r)/2)
			}
			curPoint := &point{
				X:      r,
				Y:      g,
				Z:      b,
				spread: s}
			// A color's RGBA method returns values in the range [0, 65535].
			// Shifting by 12 reduces this to the range [0, 15].
			points = append(points, curPoint)
		}
	}
	return points
}

func (p *point) String() string {
	var center string = " Center: " + p.Center.String()
	var color string = "{" + strconv.Itoa(int(p.X)) + ", " + strconv.Itoa(int(p.Y)) + ", " + strconv.Itoa(int(p.Z)) + "}"
	var str string = "Color: " + color + center
	return str
}

////////////////////////////////////////
//Utility Functions//
////////////////////////////////////////

func check(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

//getLayout will design a layout that is as square as possible Returns x width, y height
func getLayout(n int) (int, int) {
	var root float64 = math.Sqrt(float64(n))
	var h, w int = int(root), 0
	for h > 1 {
		if n%h == 0 {
			//if we find a multiple, we are done
			w = n / h
			break
		}
		//if not, try again
		h--
	}
	if h <= 1 {
		h = 1
		w = n
	}
	return w, h
}

func createColorTestImage(clusters []*cluster, fileName string, w int, h int) {
	var layoutWidth, layoutHeight int = getLayout(len(clusters))
	var pixelWidth int = layoutWidth * w
	var pixelHeight int = layoutHeight * h
	var testImage *image.RGBA = image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{pixelWidth, pixelHeight}})
	//Do rows first for memory purposes (y in outer loop)
	var r, g, b uint8 = 255, 0, 0
	for x := 0; x < pixelWidth; x++ {
		for y := 0; y < pixelHeight; y++ {
			//If we enter a new color (or a new line)
			if y%h == 0 {
				clusterNum := (y / h) + (x/w)*layoutHeight
				r = uint8(float64(clusters[clusterNum].X) / 65535 * 255)
				g = uint8(float64(clusters[clusterNum].Y) / 65535 * 255)
				b = uint8(float64(clusters[clusterNum].Z) / 65535 * 255)
			}
			c := color.RGBA{r, g, b, 255}
			testImage.Set(x, y, c)
		}
	}

	//Create the file
	myfile, _ := os.Create(fileName)
	//Write the image to the file
	png.Encode(myfile, testImage)
}

////////////////////////////////////////
//Math//
////////////////////////////////////////

func generateFirstCenters(points []*point, num int) []*cluster {
	var clusters []*cluster
	for i := 0; i < num; i++ {
		//getting the same first point is unlikely, and it will diverge even if it does happen, so do not worry about checking for it.
		//generate a random center point
		centerPoint := points[rand.Intn(len(points))]
		x, y, z := centerPoint.X, centerPoint.Y, centerPoint.Z
		cluster := &cluster{
			X:       int64(x),
			Y:       int64(y),
			Z:       int64(z),
			Members: 1}
		//TODO Bug here. If centerPoint.center != nil we should unassign it from the old one, but for our purposes it does not matter because we never use the old ones again. This is just makred in case we ever decide it needs fixed
		centerPoint.Center = cluster
		clusters = append(clusters, cluster)
	}
	return clusters
}

//sort Merge sorts the clusters based on number of members.
func clusterSort(slice []*cluster) []*cluster {
	//Base case
	if len(slice) == 1 {
		return slice
	}
	var sorted []*cluster = make([]*cluster, len(slice))
	var sortedPtr int = 0
	//Otherwise, we have some sorting to do
	//so sort each half
	left := clusterSort(slice[:len(slice)/2])
	right := clusterSort(slice[len(slice)/2:])
	leftPtr, rightPtr := 0, 0
	//and then combine them
	for i := 0; i < len(left)+len(right); i++ {
		if left[leftPtr].Members > right[rightPtr].Members {
			//left is smaller, place it in the sorted
			sorted[sortedPtr] = left[leftPtr]
			sortedPtr++
			leftPtr++
		} else {
			//right is smaller, place it in the sorted
			sorted[sortedPtr] = right[rightPtr]
			sortedPtr++
			rightPtr++
		}
		if leftPtr == len(left) {
			for _, element := range right[rightPtr:] {
				sorted[sortedPtr] = element
				sortedPtr++
			}
			break
		} else if rightPtr == len(right) {
			for _, element := range left[leftPtr:] {
				sorted[sortedPtr] = element
				sortedPtr++
			}
			break
		}
	}
	return sorted
}

//iterate will assign each point to it's closest center. it returns true if nothing changed (and we have converged)
func iterate(points []*point, clusters []*cluster) bool {
	var stabilized bool = true
	for _, point := range points {
		//find out which is the closest after last iteration.
		newClosest := point.findClosest(clusters)
		if newClosest != point.Center {
			//if it has changed,
			//we need to update the cluster TODO can we update the cluster now? I will try it and find out
			//We could have this remove done in add, but this is more clear
			point.unassignCenter()
			//Add the new point
			point.assignCenter(newClosest)
			//TODO can we get stuck ever?
			stabilized = false
		}
		//TODO DEBUG
	}
	return stabilized
}

func clusterPoints(points []*point, n int) []*cluster {
	//generate the first clusters, they will converge to actual locations later
	var clusters []*cluster = generateFirstCenters(points, n)

	for {
		//wait until convergence
		if iterate(points, clusters) {
			//nothing changed in last iteration, so we have reached a stability.
			break
		}
	}
	return clusters
}

func brightenColors(points []*point, clusters []*cluster, depth int) []*cluster {
	if depth <= 0 {
		return clusters
	}
	var brightest []*cluster = make([]*cluster, 0, len(clusters))
	for _, cluster := range clusters {
		var cPoints = make([]*point, 0, cluster.Members)
		//TODO store these somewhere so that we can skip a lot of hassle
		for _, point := range points {
			if point.Center == cluster {
				cPoints = append(cPoints, point)
			}
		}
		newClusters := clusterPoints(cPoints, len(clusters))

		//get ready to find the one with highest rgb spread
		//var brightestCluster *cluster = nil
		brightestCluster := newClusters[0]
		var brightestClusterSpread int64 = 0

		for _, newCluster := range newClusters {
			//get the min
			var min, max int64
			if newCluster.X < newCluster.Y && newCluster.X < newCluster.Z {
				min = newCluster.X
			} else if newCluster.Y < newCluster.X && newCluster.Y < newCluster.Z {
				min = newCluster.Y
			} else if newCluster.Z < newCluster.X && newCluster.Z < newCluster.Y {
				min = newCluster.Z
			}

			//Now get the max
			if newCluster.X > newCluster.Y && newCluster.X > newCluster.Z {
				max = newCluster.X
			} else if newCluster.Y > newCluster.X && newCluster.Y > newCluster.Z {
				max = newCluster.Y
			} else if newCluster.Z > newCluster.X && newCluster.Z > newCluster.Y {
				max = newCluster.Z
			}
			newCluster.colorSpread = max - min
			if newCluster.colorSpread > brightestClusterSpread {
				//new brightest
				brightestCluster = newCluster
				brightestClusterSpread = newCluster.colorSpread
			}
		}
		brightest = append(brightest, brightestCluster)
	}
	brightest = brightenColors(points, brightest, depth-1)
	return brightest
}

func main() {
	//Set up the flags
	var numClusters int
	flag.IntVar(&numClusters, "n", 10, "The number of iterations to compute")
	flag.IntVar(&numClusters, "num-clusters", 10, "The number of iterations to compute")
	var brightSteps int
	flag.IntVar(&brightSteps, "b", 1, "The number of brightening steps to calculate and give output for. A comma separated list if more than one are to be calculated")
	flag.IntVar(&brightSteps, "brightening-iterations", 1, "The number of brightening steps to calculate and give output for. A comma separated list if more than one are to be calculated")
	var outFile string
	flag.StringVar(&outFile, "out", "colorSample.png", "The name of the output file e.g. 'sample.png'")
	flag.StringVar(&outFile, "output-file", "colorSample.png", "The name of the output file e.g. 'sample.png'")
	var inFile string
	flag.StringVar(&inFile, "input-file", "", "The name of the input file e.g. 'image.png'")
	flag.StringVar(&inFile, "in", "", "The name of the input file e.g. 'image.png'")

	flag.Parse()

	if inFile == "" {
		fmt.Println("Error: No input file specified")
		flag.Usage()
		os.Exit(1)
	}

	points := getPointsFromImage(inFile)
	var clusters []*cluster = clusterPoints(points, numClusters)
	clusters = clusterSort(clusters)
	//This next step wrecks 'clusters' because all of the points are pulled out of them and put into their own. It does not matter because we never plan to use clusters again.
	brightenedClusters := brightenColors(points, clusters, brightSteps)
	createColorTestImage(brightenedClusters, outFile, 150, 75)
}
