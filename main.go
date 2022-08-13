package main

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"xml2pb/config"
	"xml2pb/log"
	"xml2pb/parsexls2bin"
	"xml2pb/parsexml"

	"github.com/tealeg/xlsx/v3"
)

var PbDefineDataList []*parsexml.PbMsgFile

func main() {
	// content, err := ioutil.ReadFile(config.ConfigData.XlsxNamePath)
	// if err != nil {
	// 	log.Log.WithField("err", err.Error()).Fatal()
	// }

	// fileList := strings.Split(string(content), "\r\n")
	// for _, v := range fileList {
	// 	fileName := strings.Split(v, " ")[0]
	// 	log.Log.WithField("file", fileName).Info()

	// }

	str := "12341"
	ss, err := regexp.MatchString("[0-9]*", str)
	if err != nil {
		return
	}
	fmt.Println(ss)
	xmlFileList, err := ioutil.ReadDir(config.ConfigData.XmlPath)
	if err != nil {
		log.Log.WithField("err", err.Error()).Fatal("Read Xml Failed")
	}

	for _, file := range xmlFileList {
		if file.IsDir() {
			continue
		}

		if strings.Contains(file.Name(), ".xml") {
			log.Log.WithField("FileName", file.Name()).Info()
			PbDefineData := new(parsexml.PbMsgFile)
			PbDefineData.Parse(config.ConfigData.XmlPath + file.Name())
			PbDefineDataList = append(PbDefineDataList, PbDefineData)
		}
	}

	xlsxFileList, err := ioutil.ReadDir(config.ConfigData.XlsxPath)
	if err != nil {
		log.Log.WithField("err", err.Error()).Fatal("Read Xlsx Failed")
	}

	for _, file := range xlsxFileList {
		if file.IsDir() {
			continue
		}

		if strings.Contains(file.Name(), ".xlsx") {
			log.Log.WithField("FileName", file.Name()).Info()
			xlsFile, err := xlsx.OpenFile(config.ConfigData.XlsxPath + file.Name())
			if err != nil {
				log.Log.WithField("err", err.Error()).Fatal("Open Xlsx Failed")
			}
			for _, sheet := range xlsFile.Sheets {
				pPbDefineData := GetPbDefineData(sheet.Name)
				if pPbDefineData == nil {
					log.Log.WithField("sheetName", sheet.Name).Fatal("xml not define")
					continue
				}

				xlsxSheetData := new(parsexls2bin.XlsxSheet)
				xlsxSheetData.Parse(sheet, pPbDefineData)
			}
		}
	}
}

func GetPbDefineData(sheetName string) *parsexml.PbMsgFile {
	for _, PbDefineData := range PbDefineDataList {
		if PbDefineData.Check(sheetName) {
			return PbDefineData
		}
	}

	return nil
}
