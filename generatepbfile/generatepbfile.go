package generatepbfile

import (
	"bufio"
	"errors"
	"os"
	"os/exec"
)

func GeneratePbFile(fileName string, context string) error {
	fileName += ".proto"
	var pbFileContext string
	pbFileContext += "/**\n"
	pbFileContext += "* @file:   " + fileName + "\n"
	pbFileContext += "* @author: peng \n"
	pbFileContext += "* @brief:  这个文件是通过工具自动生成的，建议不要手动修改\n"
	pbFileContext += "*/\n"
	pbFileContext += "syntax=\"proto3\";\n"
	pbFileContext += "option go_package=\"./resource;resource\";\n"
	pbFileContext += "package resource;\n\n"
	pbFileContext += context

	pbFile, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return errors.New("OpenFile Failed")
	}
	defer pbFile.Close()

	pbFWrite := bufio.NewWriter(pbFile)

	pbFWrite.WriteString(pbFileContext)

	pbFWrite.Flush()

	cmd := exec.Command("protoc", "--go_out=.", fileName)
	cmd.Run()

	return nil
}
