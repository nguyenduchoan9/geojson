package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

type geocsv struct {
	District, Ward, Geos, Ok string
}

type geoJson struct {
	Name         string    `json:"name"`
	Name_en      string    `json:"name_en"`
	Name_ko      string    `json:"name_ko"`
	Name_zh_hans string    `json:"name_zh_hans"`
	Name_zh_hant string    `json:"name_zh_hant"`
	Name_ja      string    `json:"name_ja"`
	Name_id      string    `json:"name_id"`
	Name_vi      string    `json:"name_vi"`
	Name_km      string    `json:"name_km"`
	MyType       string    `json:"type"`
	Features     []feature `json:"features"`
}

type feature struct {
	MyType     string     `json:"type"`
	Geometry   geometry   `json:"geometry"`
	Properties properties `json:"properties"`
}

type geometry struct {
	MyType      string        `json:"type"`
	Coordinates [][][]float64 `json:"coordinates"`
}

type coordinate [][]float64

type properties struct {
	Name string `json:"name`
}

func main() {
	filePath := flag.String("f", "", "File input")
	fdMode := flag.Int("d", 0, "0: normal/1: FD zone mode")
	flag.Parse()

	if flag.NFlag() == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}
	fb := false
	if *fdMode == 1 {
		fb = true
	}
	fmt.Printf("file %+v\n", fb)
	fmt.Printf("fd %+v\n", fdMode)
	if geojson, success := parseCSV(*filePath); success {
		filteredGeos := filter(geojson, fb)
		jsonString := toGeoJson(filteredGeos)
		toJsonFile(jsonString, fb)
	}
}

func parseCSV(fileName string) ([]geocsv, bool) {
	data := make([]geocsv, 0)
	success := true
	if csvFile, err := os.Open(fileName); err == nil {
		reader := csv.NewReader(bufio.NewReader(csvFile))
		index := 0
		for {
			if line, readErr := reader.Read(); readErr == nil {
				if index == 0 {
					index++
					continue
				} else {
					item := geocsv{
						District: line[0],
						Ward:     line[1],
						Geos:     line[2],
						Ok:       line[4],
					}
					data = append(data, item)
				}
			} else {
				if readErr == io.EOF {
					break
				} else {
					success = false
					log.Fatal(readErr)
				}
			}
		}
	} else {
		fmt.Printf("%+v\n", err)
		success = false
	}

	return data, success
}

func filter(geos []geocsv, fdZone bool) []geocsv {
	fmt.Printf("filter %+v\n", fdZone)
	filteredGeos := make([]geocsv, 0)
	for _, geo := range geos {
		if fdZone {
			if strings.EqualFold("Ok", geo.Ok) {
				filteredGeos = append(filteredGeos, geo)
			}
		} else if !strings.EqualFold("Ok", geo.Ok) {
			filteredGeos = append(filteredGeos, geo)
		}

	}
	return filteredGeos
}

func toGeoJson(geos []geocsv) geoJson {
	return geoJson{
		Name:     "VietNam",
		Name_en:  "VietNam",
		Name_vi:  "Việt Nam",
		MyType:   "FeatureCollection",
		Features: toFeature(geos),
	}
}

func groupDistrict(geos []geoJson) []geoJson {
	_geoJSON := make([]geoJson, 0)
	_geo := geoJson{Name: ""}
	for index, geo := range geos {
		if index == 0 {
			_geo = geo
			continue
		}
		priorDistrict := strings.Split(_geo.Name, "-")[0]
		currentDistrict := strings.Split(geo.Name, "-")[0]
		if priorDistrict == currentDistrict {
			_geo.Features[0].Geometry.Coordinates = append(_geo.Features[0].Geometry.Coordinates, geo.Features[0].Geometry.Coordinates[0])
		} else {
			if len(_geo.Features[0].Geometry.Coordinates) > 0 {
				_geo.Features[0].Geometry.MyType = "MultiPolygon"
			} else {
				_geo.Features[0].Geometry.MyType = "Polygon"
			}
			_geoJSON = append(_geoJSON, _geo)
			_geo = geo
		}
	}
	_geoJSON = append(_geoJSON, _geo)
	return _geoJSON
}

func toCoordinate(coorString string) [][][]float64 {
	coordinates := make([][]float64, 0)
	l := len(coorString) - 1
	coordinateStrings := strings.Split(coorString[1:l], " ")
	for _, cor := range coordinateStrings {
		coorArr := strings.Split(cor, ",")
		long, _ := strconv.ParseFloat(coorArr[0], 32)
		lat, _ := strconv.ParseFloat(coorArr[1], 32)
		coordinates = append(coordinates, []float64{long, lat})
	}
	return [][][]float64{coordinates}
}

func toJsonFile(geos geoJson, fdMode bool) {
	outputFile := "non_FD_zone.geojson"
	if fdMode {
		outputFile = "FD_zone.geojson"
	}
	if _, err := os.Stat(outputFile); err == nil {
		os.Remove(outputFile)
	}
	os.Create(outputFile)
	marsharlJson, _ := json.Marshal(geos)
	fmt.Printf("%d", len(marsharlJson))
	err := ioutil.WriteFile(outputFile, marsharlJson, os.ModeDevice)
	fmt.Println(err)
}

func toFeature(geos []geocsv) []feature {
	features := make([]feature, 0)
	for _, geo := range geos {
		featureProperties := properties{
			Name: "District " + geo.District + " - Ward " + geo.Ward,
		}

		feautureGeometry := geometry{
			MyType:      "Polygon",
			Coordinates: toCoordinate(geo.Geos),
		}
		featureType := "Feature"
		_feature := feature{
			MyType:     featureType,
			Geometry:   feautureGeometry,
			Properties: featureProperties,
		}
		features = append(features, _feature)
	}
	return features
}
