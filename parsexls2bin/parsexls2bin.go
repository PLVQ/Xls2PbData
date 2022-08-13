package parsexls2bin

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"xml2pb/log"
	"xml2pb/parsexml"

	"github.com/sirupsen/logrus"
	"github.com/tealeg/xlsx/v3"
)

const (
	FIELD_SINGULAR_TYPE = "singular"
	FIELD_REPEATED_TYPE = "repeated"
)

type XlsxSheet struct {
	dataName    string
	dataRowList []interface{}
}

func (pSheet *XlsxSheet) Parse(sheet *xlsx.Sheet, pPbDefineData *parsexml.PbMsgFile) {
	log.Log.WithFields(logrus.Fields{"sheetName": sheet.Name, "sheetRow": sheet.MaxRow, "sheetCol": sheet.MaxCol}).Info()
	pSheet.dataName = pPbDefineData.GetPbMsgName(sheet.Name)
	pSheet.dataRowList = make([]interface{}, sheet.MaxRow-1)
	pSheetPbMsg := pPbDefineData.GetPbMsg(sheet.Name)
	if pSheetPbMsg == nil {
		log.Log.WithField("sheetName", sheet.Name).Error("xml define error")
		return
	}
	for row := 1; row < sheet.MaxRow; row++ {
		// 判断xlsx行数据是否为空
		cellval, err := sheet.Cell(row, 0)
		if err != nil {
			log.Log.WithField("error", err.Error()).Error("get sheet cell error")
			break
		}
		if cellval.String() == "" {
			log.Log.WithField("row", row).Error("row begin cell is null")
			break
		}
		// 遍历行
		msg := make(map[string]interface{})
		for col := 0; col < sheet.MaxCol; col++ {
			fieldCell, err := sheet.Cell(0, col)
			if err != nil {
				log.Log.WithField("error", err.Error()).Error("get sheet field cell error")
				continue
			}

			cell, err := sheet.Cell(row, col)
			if err != nil {
				log.Log.WithFields(logrus.Fields{"row": row, "col": col, "error": err.Error()}).Error("get sheet cell error")
				continue
			}

			// 判断是否是嵌套类型
			if !strings.Contains(fieldCell.String(), "_") {
				fieldCName := fieldCell.String()
				pField := pSheetPbMsg.GetField(fieldCName)
				if pField == nil {
					continue
				}
				cellVal, err := parseCell(pField.Rule, pField.Type, cell)
				if err != nil {
					fmt.Println(err.Error())
					continue
				}
				msg[pField.Name] = cellVal
			} else {
				fieldCNameList := strings.Split(fieldCell.String(), "_")
				pField := pSheetPbMsg.GetField(fieldCNameList[0])
				var field []interface{}
				var bRepeated bool
				var fieldNameList []string
				var arrIndex int
				if pField.Rule == FIELD_REPEATED_TYPE {
					if len(fieldCNameList) < 2 {
						log.Log.WithFields(logrus.Fields{"col": col}).Fatal("sheet field cell define error")
					}
					bRepeated = true
					arrIndex, err = strconv.Atoi(fieldCNameList[1])
					if err != nil {
						log.Log.WithFields(logrus.Fields{"col": col, "error": err.Error()}).Fatal("arridx atoi failed")
					}
					field = make([]interface{}, pField.Size)
					fieldCNameList = fieldCNameList[2:]
					fieldNameList = make([]string, len(fieldCNameList))
				} else {
					fieldCNameList = fieldCNameList[1:]
					fieldNameList = make([]string, len(fieldCNameList))
				}

				pLastField := pField
				for i := 0; i < len(fieldCNameList); i++ {
					pSubPbMsg := pPbDefineData.GetPbMsgByName(pLastField.Type)
					if pSubPbMsg != nil {
						pLastField = pSubPbMsg.GetField(fieldCNameList[i])
						fieldNameList[i] = pLastField.Name
					}
				}

				cellVal, err := cellConv(pLastField.Type, cell)
				if err != nil {
					log.Log.WithFields(logrus.Fields{"row": row, "col": col, "type": pLastField.Type, "error": err.Error()}).Error("cellConv failed")
					continue
				}

				if len(fieldNameList) > 0 {
					subMsg := *parseField(fieldNameList, cellVal)
					if bRepeated {
						_, ok := msg[pField.Name]
						if !ok {
							field[arrIndex] = subMsg
							msg[pField.Name] = field
						} else {
							_, ok := msg[pField.Name].([]interface{})[arrIndex].(map[string]interface{})
							if !ok {
								msg[pField.Name].([]interface{})[arrIndex] = subMsg
								continue
							} else {
								mergeMsg(fieldNameList, msg[pField.Name].([]interface{})[arrIndex].(map[string]interface{}), subMsg)
							}
						}
					} else {
						_, ok := msg[pField.Name]
						if !ok {
							msg[pField.Name] = subMsg
							break
						} else {
							mergeMsg(fieldNameList, msg[pField.Name].(map[string]interface{}), subMsg)
						}
					}
				} else {
					_, ok := msg[pField.Name]
					if !ok {
						field[arrIndex] = cellVal
						msg[pField.Name] = field
					} else {
						msg[pField.Name].([]interface{})[arrIndex] = cellVal
					}
				}
			}
		}

		pSheet.dataRowList[row-1] = msg
	}
	pSheet.GenerateBin()
}

func parseField(fieldNameList []string, fieldValue interface{}) *map[string]interface{} {
	iLen := len(fieldNameList)
	fieldStruck := make(map[string]interface{})
	fieldStruck[fieldNameList[iLen-1]] = fieldValue
	// feildInfo := new(map[string]interface{})
	feildInfo := &fieldStruck
	for i := iLen - 2; i >= 0; i-- {
		tempField := make(map[string]interface{})
		tempField[fieldNameList[i]], fieldStruck = fieldStruck, tempField
		feildInfo = &tempField
	}
	return feildInfo
}

func mergeMsg(fieldNameList []string, destMsg map[string]interface{}, srcMsg map[string]interface{}) {
	for i := 0; i < len(fieldNameList); i++ {
		_, ok := destMsg[fieldNameList[i]]
		if !ok {
			destMsg[fieldNameList[i]] = srcMsg[fieldNameList[i]]
			break
		}
		destMsg = destMsg[fieldNameList[i]].(map[string]interface{})
		srcMsg = srcMsg[fieldNameList[i]].(map[string]interface{})
	}
}

func (pSheet *XlsxSheet) GenerateBin() {
	fmt.Println(pSheet.dataRowList)
	mm := map[string]interface{}{"data": pSheet.dataRowList}
	data, err := json.Marshal(mm)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	log.Log.WithField("data", string(data)).Info()
	binName := pSheet.dataName + ".bin"
	pbFile, err := os.OpenFile(binName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	defer pbFile.Close()

	pbFWrite := bufio.NewWriter(pbFile)

	pbFWrite.Write(data)

	pbFWrite.Flush()
}

func parseCell(fieldRule string, fieldType string, cell *xlsx.Cell) (interface{}, error) {
	if fieldRule == FIELD_REPEATED_TYPE {
		cellValList := strings.Split(cell.String(), "|")
		if fieldType == "string" {
			return cellValList, nil
		} else if fieldType == "int64" {
			realValList := make([]int64, len(cellValList))
			for idx, cellVal := range cellValList {
				val, err := strconv.Atoi(cellVal)
				if err != nil {
					return nil, err
				}

				realValList[idx] = int64(val)
			}
			return realValList, nil
		} else if fieldType == "float64" {
			realValList := make([]float64, len(cellValList))
			for idx, cellVal := range cellValList {
				val, err := strconv.ParseFloat(cellVal, 64)
				if err != nil {
					return nil, err
				}

				realValList[idx] = val
			}
			return realValList, nil
		}

	} else {
		return cellConv(fieldType, cell)
	}

	return nil, errors.New("parse err")
}

func cellConv(fieldType string, cell *xlsx.Cell) (interface{}, error) {
	switch fieldType {
	case "double":
		cellVal, err := cell.Float()
		return float64(cellVal), err
	case "int32":
		cellVal, err := cell.Int64()
		return int32(cellVal), err
	case "uint32":
		cellVal, err := cell.Int64()
		return uint32(cellVal), err
	case "sint32":
		cellVal, err := cell.Int64()
		return int32(cellVal), err
	case "fixed32":
		cellVal, err := cell.Int64()
		return uint32(cellVal), err
	case "sfixed32":
		cellVal, err := cell.Int64()
		return int32(cellVal), err
	case "bool":
		cellVal := cell.Bool()
		return bool(cellVal), nil
	case "float":
		cellVal, err := cell.Float()
		return float32(cellVal), err
	case "int64":
		cellVal, err := cell.Int64()
		return int64(cellVal), err
	case "uint64":
		cellVal, err := cell.Int64()
		return uint64(cellVal), err
	case "sint64":
		cellVal, err := cell.Int64()
		return int64(cellVal), err
	case "fixed64":
		cellVal, err := cell.Int64()
		return uint64(cellVal), err
	case "sfixed64":
		cellVal, err := cell.Int64()
		return int64(cellVal), err
	case "string":
		return cell.String(), nil
	default:
		return nil, nil
	}
}
