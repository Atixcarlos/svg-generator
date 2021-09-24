package main

import (
    "encoding/base64"
    "encoding/json"
	"github.com/ajstarks/svgo/float"
	"io/ioutil"
    "os"
    "fmt"
    "bytes"
    "sort"
	"flag"
    "crypto/rand"
)

type WallComplex struct {
    Exposure    string
    WallLength float64
    Automatic bool
}

// randcolor returns a random color
func randcolor() string {
	rgb := []byte{0, 0, 0} // read error returns black
	rand.Read(rgb)
	return fmt.Sprintf("rgb(%d,%d,%d)", rgb[0], rgb[1], rgb[2])
}

func indexOf(element string, data []string) (int) {
   for k, v := range data {
       if element == v {
           return k
       }
   }
   return -1 //not found.
}

func isExists(exposure string, walls []WallComplex) (result bool) {
	result = false
	for _, w := range walls {
		if w.Exposure == exposure {
			result = true
			break
		}
	}
	return result
}

func CompleteExposures(myIncompleteExposures []WallComplex) ([]WallComplex) {
    var myCompleteExposures []WallComplex

    // Define required exposure with their opposites.     
    myRequiredExposures :=  map[string]string{
        "N": "S",
        "E": "W",
        "S": "N", 
        "W": "E",
    }

    for _, w := range myIncompleteExposures { 
        myCompleteExposures = append(myCompleteExposures, WallComplex{w.Exposure, w.WallLength, false})
        myOposite := myRequiredExposures[w.Exposure]
        if !isExists(myOposite, myIncompleteExposures) {
            myCompleteExposures = append(myCompleteExposures, WallComplex{myOposite, w.WallLength, true})
        }
    }

    for _, w := range myIncompleteExposures { 
        var matchingExposures []string
        switch w.Exposure {
            case "N": matchingExposures = []string{ "E", "W" }
            case "E": matchingExposures = []string{ "N", "S" }
            case "S": matchingExposures = []string{ "E", "W" }
            case "W": matchingExposures = []string{ "N", "S" }
        }
        for _, matchingExposure := range matchingExposures {
            if !isExists(matchingExposure, myCompleteExposures) {
                myCompleteExposures = append(myCompleteExposures, WallComplex{matchingExposure, w.WallLength, true})
            }
        }
    }
    return myCompleteExposures  
}

func createShapes(exposures []WallComplex, shape string, s *svg.SVG) {
    // StartPoint    
    myCenterPointX := 200.0
    myCenterPointY := 200.0

    var myRoomPointsX []float64 
    var myRoomPointsY []float64 
      
    // Sort slice to add every wall in order to exposure.
    myOrderedExposures :=  []string{"N", "NE", "E", "SE", "S", "SW", "W", "NW"}
    sort.Slice(exposures, func(i, j int) bool {
        return indexOf(exposures[i].Exposure, myOrderedExposures) < indexOf(exposures[j].Exposure, myOrderedExposures)
    })

    for _, w := range exposures { 
        x1 := myCenterPointX 
        y1 := myCenterPointY
        x2 := myCenterPointX 
        y2 := myCenterPointY  

       switch exp := w.Exposure; exp {
	    case "N":
            x2 = myCenterPointX + w.WallLength  
         
        case "NE": 
            x2 = myCenterPointX + w.WallLength  
            y2 = myCenterPointY + w.WallLength
        
        case "E": 
            y2 = myCenterPointY + w.WallLength
        
        case "SE": 
            x2 = myCenterPointX - w.WallLength  
            y2 = myCenterPointY + w.WallLength
       
        case "S":
            x2 = myCenterPointX - w.WallLength
        
        case "SW": 
            x2 = myCenterPointX - w.WallLength  
            y2 = myCenterPointY - w.WallLength
        
        case "W":
            y2 = myCenterPointY - w.WallLength
         
        case "NW": 
            x2 = myCenterPointX + w.WallLength  
            y2 = myCenterPointY - w.WallLength
	    }

        // Set Center point with the last point of the last wall drawed.
        myCenterPointX = x2
        myCenterPointY = y2 

        // Draw line based points
        if shape == "Line" {
            if !w.Automatic {
                s.Line(x1, y1, x2, y2, "fill:none; stroke-width:5; stroke:"+randcolor())
            }
        }
        
        if shape == "Polygon" { 
            myRoomPointsX = append(myRoomPointsX,x2)
            myRoomPointsY = append(myRoomPointsY,y2)
        }
    }

    if shape == "Polygon" {
        s.Polygon(myRoomPointsX, myRoomPointsY, "fill:#EFEFEF")
    }
}

func main() {
    // Get walls data from json file
	// Read parameters from terminal (json file).
 	myJsonFile := flag.String("json", "my_walls.json", "json file")
	flag.Parse()

	jsonFile, err := os.Open(*myJsonFile)
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()
	myJsonBytes, _ := ioutil.ReadAll(jsonFile)

	// Decode json to walls variable
	var myWalls []WallComplex
	json.Unmarshal(myJsonBytes, &myWalls)

    //Create a buffer to save all the output
    buf := new(bytes.Buffer)
    s := svg.New(buf)
    //s.Startview(400.00, 400.00, -10.00, 20.00, 300.00, 300.00)
    s.Startview(400.00, 400.00, 0.00, 0.00, 600.00, 600.00)
    
    // create a group with specified id    
    s.Gid("1")

    myCompleteExposures := CompleteExposures(myWalls)
    createShapes(myCompleteExposures, "Line", s)

    exposuresByQuadrant := make(map[string]float64)
    for _, moreData := range myWalls { 
        if _, ok := exposuresByQuadrant[moreData.Exposure]; ok {
            exposuresByQuadrant[moreData.Exposure] += moreData.WallLength
        } else {
            exposuresByQuadrant[moreData.Exposure] = moreData.WallLength
        }
    }

    var myWallsByQuadrant []WallComplex
    for exp, val := range exposuresByQuadrant {
        myWallsByQuadrant = append(myWallsByQuadrant, WallComplex{exp, val, false})
    }

    myCompleteExposures = CompleteExposures(myWallsByQuadrant)
    createShapes(myCompleteExposures, "Polygon", s)

    s.Gend()
    s.End()
    
    // Print svg as string
    fmt.Println(buf.String())

    // Write in a file
    f, err := os.Create("out.svg")
    if err != nil {
        fmt.Println(err)
    }
    defer f.Close()
    _, err2 := f.WriteString(buf.String())
    if err2 != nil {
        fmt.Println(err2)
    }

    //svg as base64 
    sEnc := base64.StdEncoding.EncodeToString(buf.Bytes())
    fmt.Println(sEnc)
}
