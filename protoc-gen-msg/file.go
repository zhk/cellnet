package main

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"text/template"

	"github.com/davyxu/cellnet/util"
	"github.com/davyxu/pbmeta"
)

const codeTemplate = `// Generated by github.com/davyxu/cellnet/protoc-gen-msg
// DO NOT EDIT!{{range .Protos}}
// Source: {{.Name}}{{end}}

package gamedef

{{if gt .TotalMessages 0}}
import (
	"github.com/davyxu/cellnet"
)
{{end}}

func init() {
	{{range .Protos}}
	// {{.Name}}{{range .Messages}}
	cellnet.RegisterMessageMeta("{{.FullName}}", (*{{.Name}})(nil), {{.MsgID}})	{{end}}
	{{end}}
}

`

type msgModel struct {
	*pbmeta.Descriptor

	parent *pbmeta.FileDescriptor
}

func (self *msgModel) MsgID() int {
	return int(util.StringHash(self.FullName()))
}

func (self *msgModel) FullName() string {
	return fmt.Sprintf("%s.%s", self.parent.PackageName(), self.Name())
}

type protoModel struct {
	*pbmeta.FileDescriptor

	Messages []*msgModel
}

func (self *protoModel) Name() string {
	return self.FileDescriptor.FileName()
}

type fileModel struct {
	TotalMessages int
	Protos        []*protoModel
}

func printFile(pool *pbmeta.DescriptorPool) (string, bool) {

	tpl, err := template.New("msgid").Parse(codeTemplate)
	if err != nil {
		log.Errorln(err)
		return "", false
	}

	var model fileModel

	for f := 0; f < pool.FileCount(); f++ {

		file := pool.File(f)

		pm := &protoModel{
			FileDescriptor: file,
		}

		for m := 0; m < file.MessageCount(); m++ {

			d := file.Message(m)

			pm.Messages = append(pm.Messages, &msgModel{
				Descriptor: d,
				parent:     file,
			})

		}

		model.TotalMessages += file.MessageCount()

		model.Protos = append(model.Protos, pm)

	}

	var bf bytes.Buffer

	err = tpl.Execute(&bf, &model)
	if err != nil {
		log.Errorln(err)
		return "", false
	}

	err = formatCode(&bf)
	if err != nil {
		log.Errorln(err)
		return "", false
	}

	return bf.String(), true
}

func formatCode(bf *bytes.Buffer) error {
	// Reformat generated code.
	fset := token.NewFileSet()

	ast, err := parser.ParseFile(fset, "", bf, parser.ParseComments)
	if err != nil {
		return err
	}

	bf.Reset()

	err = (&printer.Config{Mode: printer.TabIndent | printer.UseSpaces, Tabwidth: 8}).Fprint(bf, fset, ast)
	if err != nil {
		return err
	}

	return nil
}
