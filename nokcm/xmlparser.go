/*
Copyright © 2020 Zhengwei Gao<zhengwei.gao@yahoo.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package nokcm

import (
	"fmt"
	"github.com/beevik/etree"
	"github.com/zhenggao2/ngapp/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"strings"
)

type XmlParser struct {
	log *zap.Logger
	out string
	debug bool
}

func (p *XmlParser) Init(log *zap.Logger, out string, debug bool) {
	p.log = log
	p.out = out
	p.debug = debug

	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Initializing XML parser..."))
}

func (p *XmlParser) Parse(xml, txml string) {
	switch strings.ToLower(txml) {
	case "scfc", "vendor":
		p.ParseScfcVendor(xml)
	case "freqhist":
		p.ParseFreqHist(xml)
	}
}

func (p *XmlParser) ParseScfcVendor(xml string) {
	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Parsing SCFC/Vendor...[%s]", xml))

	doc := etree.NewDocument()
	if err := doc.ReadFromFile(xml); err != nil {
		p.writeLog(zapcore.ErrorLevel, fmt.Sprintf("Fail to open file: %s", xml))
		return
	}

	root := doc.SelectElement("raml")
	if root == nil {
		p.writeLog(zapcore.DebugLevel, fmt.Sprintf("No raml element, xml=[%s]", xml))
		return
	}
	//fmt.Printf("[%s]: ns=%v, path=%v, index=%v, tag=%v, attr=%v\n", filepath.Base(xml), root.NamespaceURI(), root.GetPath(), root.Index(), root.Tag, root.Attr)
	// root: tag=raml, attr=[{ version 2.1 0xc000686120} { xmlns raml21.xsd 0xc000686120}]

	cm := root.SelectElement("cmData")
	if cm == nil {
		p.writeLog(zapcore.DebugLevel, fmt.Sprintf("No cmData element, xml=[%s]", xml))
		return
	}
	//fmt.Printf("[%s]: ns=%v, path=%v, index=%v, tag=%v, attr=%v\n", filepath.Base(xml), cm.NamespaceURI(), cm.GetPath(), cm.Index(), cm.Tag, cm.Attr)
	// cmData: tag=cmData, attr=[{ scope all 0xc0006861e0} { type actual 0xc0006861e0}]

	data := make(map[string]*utils.OrderedMap)
	for _, mo := range cm.FindElements("managedObject") {
		dn := mo.SelectAttrValue("distName", "")
		if len(dn) == 0 {
			p.writeLog(zapcore.DebugLevel, fmt.Sprintf("No distName attribute in managedObject element [path=%v,index=%v], xml=[%s]", mo.GetPath(), mo.Index(), xml))
			break
		}

		dn = strings.Replace(dn, "PLMN-PLMN/", "", -1)
		data[dn] = utils.NewOrderedMap()
		for _, list := range mo.FindElements("list") {
			listName := list.SelectAttrValue("name", "")

			/* --special handling of list without <item> Element
			<list name="nrhoirDNList">
				<p>MRBTS-1619304/NRBTS-1619304/NRHOIR-1</p>
				<p>MRBTS-1619304/NRBTS-1619304/NRHOIR-2</p>
				<p>MRBTS-1619304/NRBTS-1619304/NRHOIR-3</p>
				<p>MRBTS-1619304/NRBTS-1619304/NRHOIR-4</p>
			</list>
			 */
			numItem := len(list.FindElements("item"))
			if numItem == 0 {
				val := make([]string, 0)
				for _, p := range list.FindElements("p") {
					val = append(val, p.Text())
				}
				data[dn].Add(listName, val)
			} else {
				// first pass, find all fields
				fields := make(map[string]bool)
				for _, item := range list.FindElements("item") {
					for _, p := range item.FindElements("p") {
						par := listName + "." + p.SelectAttrValue("name", "")
						if _, exist := fields[par]; !exist {
							fields[par] = false
						}
					}
				}

				// second pass, update data[dn]
				for _, item := range list.FindElements("item") {
					for _, p := range item.FindElements("p") {
						par := listName + "." + p.SelectAttrValue("name", "")
						if data[dn].Exist(par) {
							data[dn].Add(par, append(data[dn].Val(par).([]string), p.Text()))
						} else {
							data[dn].Add(par, []string{p.Text()})
						}
						fields[par] = true
					}

					for par := range fields {
						if !fields[par] {
							if data[dn].Exist(par) {
								data[dn].Add(par, append(data[dn].Val(par).([]string), "-"))
							} else {
								data[dn].Add(par, []string{"-"})
							}
						} else {
							fields[par] = false
						}
					}
				}
			}
		}

		for _, p := range mo.FindElements("p") {
			par := p.SelectAttrValue("name", "")
			data[dn].Add(par, p.Text())
		}
	}

	xmlBn := filepath.Base(xml)
	ofn := filepath.Join(p.out, fmt.Sprintf("%s.dat", xmlBn[:len(xmlBn)-len(filepath.Ext(xml))]))
	fout, err := os.OpenFile(ofn, os.O_WRONLY|os.O_CREATE, 0664)
	if err != nil {
		p.writeLog(zapcore.ErrorLevel, fmt.Sprintf("Fail to open file: %s", ofn))
		return
	}

	fout.WriteString("# [dn===*]\n# name===value\n")
	for dn := range data {
		fout.WriteString(fmt.Sprintf("\n[dn===%s]\n", dn))
		for _, par := range data[dn].Keys() {
			fout.WriteString(fmt.Sprintf("%s===%v\n", par, data[dn].Val(par)))
			/*
			if _, ok := data[dn].Val(par).([]string); ok {
				fout.WriteString(fmt.Sprintf("%s===%+q\n", par, data[dn].Val(par)))
			} else {
				fout.WriteString(fmt.Sprintf("%s===%v\n", par, data[dn].Val(par)))
			}
			 */
		}
	}
	fout.Close()
}

func (p *XmlParser) ParseFreqHist(xml string) {
	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Parsing FrquencyHistory...[%s]", xml))

}

func (p *XmlParser) writeLog(level zapcore.Level, s string) {
	switch level {
	case zapcore.DebugLevel:
		p.log.Debug(s)
	case zapcore.InfoLevel:
		p.log.Info(s)
	case zapcore.WarnLevel:
		p.log.Warn(s)
	case zapcore.ErrorLevel:
		p.log.Error(s)
	case zapcore.FatalLevel:
		p.log.Fatal(s)
	case zapcore.PanicLevel:
		p.log.Panic(s)
	default:
	}

	if level != zapcore.DebugLevel {
		fmt.Println(s)
	}
}
