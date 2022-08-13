package parsexml

import (
	"encoding/xml"
	"io/ioutil"
	"xml2pb/log"
)

type PbMsgFile struct {
	ResConfig xml.Name `xml:"ResConfig"`
	Name      string   `xml:"name,attr"`
	MsgList   []PbMsg  `xml:"struct"`
}

type PbMsg struct {
	Msg       xml.Name     `xml:"struct"`
	Name      string       `xml:"name,attr"`
	CName     string       `xml:"cname,attr"`
	Desc      string       `xml:"desc,attr"`
	FieldList []PbMsgField `xml:"entry"`
}

type PbMsgField struct {
	Field xml.Name `xml:"entry"`
	Name  string   `xml:"name,attr"`
	Rule  string   `xml:"rule,attr"`
	Type  string   `xml:"type,attr"`
	Size  int      `xml:"size,attr"`
	CName string   `xml:"cname,attr"`
	Desc  string   `xml:"desc,attr"`
}

func (pPbDefineData *PbMsgFile) Parse(filePath string) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Log.WithField("err", err.Error()).Info("读文件出错！")
		return
	}
	// fmt.Println(string(bytes))
	err = xml.Unmarshal(data, pPbDefineData)
	if err != nil {
		log.Log.WithField("err", err.Error()).Info("解析文件出错！")
		return
	}
	log.Log.WithField("PbDefineData", *pPbDefineData).Info()
}

func (pPbDefineData *PbMsgFile) Check(cname string) bool {
	for _, msg := range pPbDefineData.MsgList {
		if msg.CName == cname {
			return true
		}
	}

	return false
}

func (pPbDefineData *PbMsgFile) GetPbMsgName(cname string) string {
	for _, msg := range pPbDefineData.MsgList {
		if msg.CName == cname {
			return msg.Name
		}
	}

	return ""
}

func (pPbDefineData *PbMsgFile) GetPbMsg(cname string) *PbMsg {
	for _, msg := range pPbDefineData.MsgList {
		if msg.CName == cname {
			return &msg
		}
	}

	return nil
}

func (pPbDefineData *PbMsgFile) GetPbMsgByName(name string) *PbMsg {
	for _, msg := range pPbDefineData.MsgList {
		if msg.Name == name {
			return &msg
		}
	}

	return nil
}

func (pPbMsg *PbMsg) GetField(cname string) *PbMsgField {

	for _, field := range pPbMsg.FieldList {
		if field.CName == cname {
			return &field
		}
	}

	return nil
}
