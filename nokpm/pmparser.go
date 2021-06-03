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
package nokpm

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/beevik/etree"
	"github.com/zhenggao2/ngapp/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path"
	"strings"
)

// keyPerAgg is map between measurement aggregation and its default key pattern in PM database.
var keyPerAgg = map[string]string {
	"NRDU" : "NRBTSID_NRDUID_TS",
	"NRCUUP" : "NRBTSID_NRCUUPID_TS",
	"NRBTS" : "NRBTSID_TS",
	"NRBTS_PLMN" : "NRBTSID_MCC_MNC_TS",
	"NRBTS_PLMN_SLICE" : "NRBTSID_MCC_MNC_SST_SD_TS",
	"NRCELL" : "NRBTSID_NRCELLID_TS",
	"NRCELL_PLMN" : "NRBTSID_NRCELLID_MCC_MNC_TS",
	"NRCELL_NRREL" : "NRBTSID_NRCELLID_NRCI_MCC_MNC_TS",
	"NRCELL_NRRELE" : "NRBTSID_NRCELLID_ECI_MCC_MNC_TS",
}

// aggPerMeas is map between measurement type and its aggregation.
var aggPerMeasType = map[string]string {
	"NX2CB" : "NRBTS",
	"NX2CC" : "NRCELL",
	"NF1CC" : "NRCELL",
	"RACCU" : "NRCELL",
	"RACDU" : "NRCELL",
	"NNSAU" : "NRCELL",
	"NF1CD" : "NRDU",
	"NCAC" : "NRCELL",
	"NNGCB" : "NRBTS",
	"NINFC" : "NRCELL",
	"NE1CB" : "NRBTS",
	"NE1CU" : "NRCUUP",
	"NRRCC" : "NRCELL",
	"NNGCC" : "NRCELL",
	"NXNCB" : "NRBTS",
	"NRRCD" : "NRDU",
	"LTEMO" : "NRBTS",
	"NRLFC" : "NRCELL",
	"SRB3C" : "NRCELL",
	"NTRAF" : "NRBTS",
	"NE1UN" : "NRCUUP",
	"NE1US" : "NRCUUP",
	"NE1CN" : "NRBTS",
	"NE1CS" : "NRBTS",
	"LTEMG" : "NRCELL",
	"RSACU" : "NRCELL",
	"RSADU" : "NRCELL",
	"RSANB" : "NRBTS",
	"SRB3D" : "NRCELL",
	"NXNCC" : "NRCELL",
	"NRANS" : "NRCELL_PLMN",
	"NRBC" : "NRCELL",
	"NMOCU" : "NRCELL",
	"NMODU" : "NRCELL",
	"NSL" : "NRBTS_PLMN_SLICE",
	"NF1UU" : "NRDU",
	"NIFC" : "NRCELL",
	"NRMG" : "NRCELL",
	"NRREL" : "NRCELL_NRREL",
	"NGCFB" : "NRCELL", // FIXME
	"NDLHQ" : "NRCELL",
	"NDLSQ" : "NRCELL",
	"NULHQ" : "NRCELL",
	"NULSQ" : "NRCELL",
	"NMPDU" : "NRCELL",
	"NRBF" : "NRCELL",
	"NRACH" : "NRCELL",
	"NRTA" : "NRCELL",
	"NCELA" : "NRCELL",
	"NMSDU" : "NRCELL",
	"NLRLC" : "NRCELL",
	"NHRLC" : "NRDU",
	"NX2UB" : "NRBTS",
	"NF1UD" : "NRDU",
	"NS1UB" : "NRBTS",
	"NCAD" : "NRCELL",
	"NPDCC" : "NRBTS",
	"NECUP" : "NRCELL",
	"PDCP1" : "NRBTS",
	"PDCP2" : "NRBTS",
	"NF1UB" : "NRBTS",
	"NEIRP" : "NRCELL",
	"NRASU" : "NRBTS_PLMN",
	"NRNGU" : "NRBTS",
	"NRXNU" : "NRBTS",
	"NRPAG" : "NRCELL",
	"ENDSS" : "NRCELL",
	"PDCCH" : "NRCELL",
	"NCAV" : "NRCELL",
	"NREMO" : "NRCELL_NRRELE", // FIXME, new in 5G21A
	"NPSL" : "NRBTS_PLMN_SLICE", // new in 5G21A
}

// measId2MeasType is map between measurement ID and its type.
var measId2MeasType = map[string]string {
	"M55110" : "NX2CB",
	"M55112" : "NX2CC",
	"M55113" : "NF1CC",
	"M55114" : "RACCU",
	"M55115" : "RACDU",
	"M55116" : "NNSAU",
	"M55117" : "NF1CD",
	"M55118" : "NCAC",
	"M55119" : "NNGCB",
	"M55120" : "NINFC",
	"M55121" : "NE1CB",
	"M55123" : "NE1CU",
	"M55124" : "NRRCC",
	"M55125" : "NNGCC",
	"M55126" : "NXNCB",
	"M55127" : "NRRCD",
	"M55128" : "LTEMO",
	"M55129" : "NRLFC",
	"M55130" : "SRB3C",
	"M55131" : "NTRAF",
	"M55132" : "NE1UN",
	"M55133" : "NE1US",
	"M55134" : "NE1CN",
	"M55135" : "NE1CS",
	"M55136" : "LTEMG",
	"M55138" : "RSACU",
	"M55139" : "RSADU",
	"M55140" : "RSANB",
	"M55141" : "SRB3D",
	"M55143" : "NXNCC",
	"M55145" : "NRANS",
	"M55146" : "NRBC",
	"M55147" : "NMOCU",
	"M55148" : "NMODU",
	"M55149" : "NSL",
	"M55150" : "NF1UU",
	"M55151" : "NIFC",
	"M55152" : "NRMG",
	"M55153" : "NRREL",
	"M55155" : "NGCFB",
	"M55300" : "NDLHQ",
	"M55301" : "NDLSQ",
	"M55302" : "NULHQ",
	"M55303" : "NULSQ",
	"M55304" : "NMPDU",
	"M55305" : "NRBF",
	"M55306" : "NRACH",
	"M55307" : "NRTA",
	"M55308" : "NCELA",
	"M55309" : "NMSDU",
	"M55310" : "NLRLC",
	"M55311" : "NHRLC",
	"M55313" : "NX2UB",
	"M55314" : "NF1UD",
	"M55315" : "NS1UB",
	"M55316" : "NCAD",
	"M55317" : "NPDCC",
	"M55318" : "NECUP",
	"M55319" : "PDCP1",
	"M55320" : "PDCP2",
	"M55323" : "NF1UB",
	"M55324" : "NEIRP",
	"M55326" : "NRASU",
	"M55327" : "NRNGU",
	"M55329" : "NRXNU",
	"M55331" : "NRPAG",
	"M55332" : "ENDSS",
	"M55335" : "PDCCH",
	"M55601" : "NCAV",
	"M55602" : "SBM",
	"M55603" : "SFPRM",
	"M55800" : "NGNS",
	"M55144" : "NREMO",
	"M55328" : "NPSL",
}

// measType2MeasId is map between measurement type and its ID.
var measType2MeasId = map[string]string {
	"NX2CB" : "M55110",
	"NX2CC" : "M55112",
	"NF1CC" : "M55113",
	"RACCU" : "M55114",
	"RACDU" : "M55115",
	"NNSAU" : "M55116",
	"NF1CD" : "M55117",
	"NCAC" : "M55118",
	"NNGCB" : "M55119",
	"NINFC" : "M55120",
	"NE1CB" : "M55121",
	"NE1CU" : "M55123",
	"NRRCC" : "M55124",
	"NNGCC" : "M55125",
	"NXNCB" : "M55126",
	"NRRCD" : "M55127",
	"LTEMO" : "M55128",
	"NRLFC" : "M55129",
	"SRB3C" : "M55130",
	"NTRAF" : "M55131",
	"NE1UN" : "M55132",
	"NE1US" : "M55133",
	"NE1CN" : "M55134",
	"NE1CS" : "M55135",
	"LTEMG" : "M55136",
	"RSACU" : "M55138",
	"RSADU" : "M55139",
	"RSANB" : "M55140",
	"SRB3D" : "M55141",
	"NXNCC" : "M55143",
	"NRANS" : "M55145",
	"NRBC" : "M55146",
	"NMOCU" : "M55147",
	"NMODU" : "M55148",
	"NSL" : "M55149",
	"NF1UU" : "M55150",
	"NIFC" : "M55151",
	"NRMG" : "M55152",
	"NRREL" : "M55153",
	"NGCFB" : "M55155",
	"NDLHQ" : "M55300",
	"NDLSQ" : "M55301",
	"NULHQ" : "M55302",
	"NULSQ" : "M55303",
	"NMPDU" : "M55304",
	"NRBF" : "M55305",
	"NRACH" : "M55306",
	"NRTA" : "M55307",
	"NCELA" : "M55308",
	"NMSDU" : "M55309",
	"NLRLC" : "M55310",
	"NHRLC" : "M55311",
	"NX2UB" : "M55313",
	"NF1UD" : "M55314",
	"NS1UB" : "M55315",
	"NCAD" : "M55316",
	"NPDCC" : "M55317",
	"NECUP" : "M55318",
	"PDCP1" : "M55319",
	"PDCP2" : "M55320",
	"NF1UB" : "M55323",
	"NEIRP" : "M55324",
	"NRASU" : "M55326",
	"NRNGU" : "M55327",
	"NRXNU" : "M55329",
	"NRPAG" : "M55331",
	"ENDSS" : "M55332",
	"PDCCH" : "M55335",
	"NCAV" : "M55601",
	"SBM" : "M55602",
	"SFPRM" : "M55603",
	"NGNS" : "M55800",
	"NREMO" : "M55144",
	"NPSL" : "M55328",
}

// keyPatTwmXinos is map between token of the default key pattern and the field of TWM XINOS.
var keyPatTwmXinos = map[string]string {
	"NRBTSID" : "NRBTS_ID",
	"NRCELLID" : "NRCELL_ID",
	"NRDUID" : "NRDU",
	"NRCUUPID" : "NRCUUP",
	"MCC" : "MCC",
	"MNC" : "MNC",
	"SST" : "SST",
	"SD" : "SD",
	"NRCI" : "NRCI",
	"ECI" : "ECI",
	"TS" : "TIME",
}

type PmParser struct {
	log   *zap.Logger
	op    string
	db    string
	debug bool
}

type CsvHeaderPos struct {
	keyPatPos map[string]int
	posStart int // position of first counter
}

func (p *PmParser) Init(log *zap.Logger, op, db string, debug bool) {
	p.log = log
	p.op = op
	p.db = db
	p.debug = debug

	// For TWM XINOS M55145(NRANS), the aggregation is NRBTS_PLMN
	if p.op == "twm" {
		aggPerMeasType["NRANS"] = "NRBTS_PLMN"
	}

	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Initializing PM parser..."))
}

func (p *PmParser) Parse(pm, tpm string) {
	switch strings.ToLower(tpm) {
	case "raw":
		p.ParseRawPmXml(pm)
	case "sql":
		p.ParseSqlQueryCsv(pm)
	}
}

func (p *PmParser) ParseRawPmXml(xml string) {
	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Parsing raw PM...[%s]", xml))

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
	//fmt.Printf("[%s]: ns=%v, path=%v, index=%v, tag=%v, attr=%v\n", path.Base(xml), root.NamespaceURI(), root.GetPath(), root.Index(), root.Tag, root.Attr)
	// root: tag=raml, attr=[{ version 2.1 0xc000686120} { xmlns raml21.xsd 0xc000686120}]

	cm := root.SelectElement("cmData")
	if cm == nil {
		p.writeLog(zapcore.DebugLevel, fmt.Sprintf("No cmData element, xml=[%s]", xml))
		return
	}
	//fmt.Printf("[%s]: ns=%v, path=%v, index=%v, tag=%v, attr=%v\n", path.Base(xml), cm.NamespaceURI(), cm.GetPath(), cm.Index(), cm.Tag, cm.Attr)
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

		for _, p := range mo.FindElements("p") {
			par := p.SelectAttrValue("name", "")
			data[dn].Add(par, p.Text())
		}
	}

	xmlBn := path.Base(xml)
	ofn := path.Join(p.db, fmt.Sprintf("%s.dat", xmlBn[:len(xmlBn)-len(path.Ext(xml))]))
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

func (p *PmParser) ParseSqlQueryCsv(csv string) {
	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Parsing csv PM(operator:%s)...[%s]", p.op, path.Base(csv)))

	fin, err := os.Open(csv)
	if err != nil {
		p.writeLog(zapcore.ErrorLevel, err.Error())
		return
	}

	reader := bufio.NewReader(fin)

	// read header
	header, _ := reader.ReadString('\n')
	// remove preceding 0xefbbbf and ending 0x0A(0x0A0D return and newline) of BOM format
	boom := []byte{0xEF, 0xBB, 0xBF, 0x0A}
	header= strings.Trim(header, string(boom))
	headerTokens := strings.Split(header, ",")
	var pos CsvHeaderPos
	if p.op == "twm" {
		pos = p.findKeyPatPosTwmXinos(headerTokens)
	}
	measType, exist := measId2MeasType[strings.Split(headerTokens[pos.posStart], "C")[0]]
	if !exist {
		p.writeLog(zapcore.FatalLevel, fmt.Sprintf("Unsupported measurement type! Counter name is %s.", headerTokens[pos.posStart]))
		fin.Close()
		return
	}

	data := make(map[string]*bytes.Buffer)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		// remove leading and tailing spaces
		line = strings.TrimSpace(line)
		tokens := strings.Split(line, ",")

		keyPatTokens := strings.Split(keyPerAgg[aggPerMeasType[measType]], "_")
		keyTokens := make([]string, 0)
		//fmt.Printf("measType=%s,keyPatTokens=%v,p.pos=%v,headerTokens[0]=%v\n", measType, keyPatTokens, p.pos, headerTokens[0])
		for _, token := range keyPatTokens {
			keyTokens = append(keyTokens, tokens[pos.keyPatPos[token]])
		}
		key := strings.Join(keyTokens, "_")

		for i := pos.posStart; i < len(tokens); i++ {
			counter := headerTokens[i]
			val := tokens[i]
			if _, exist := data[counter]; !exist {
				data[counter] = new(bytes.Buffer)
			}
			data[counter].WriteString(key + "," + val + "\n")
		}
	}

	fin.Close()

	for counter := range data {
		ofn := path.Join(p.db, fmt.Sprintf("%s.gz", counter))
		fout, err := os.OpenFile(ofn, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0664)
		if err != nil {
			p.writeLog(zapcore.ErrorLevel, fmt.Sprintf("Fail to open file: %s", ofn))
			break
		}

		gw := gzip.NewWriter(fout)
		gw.Write(data[counter].Bytes())
		gw.Close()

		fout.Close()
	}
}

func (p *PmParser) findKeyPatPosTwmXinos(tokens []string) CsvHeaderPos {
	pos := CsvHeaderPos{
		keyPatPos: make(map[string]int),
		posStart: -1,
	}
	for k := range keyPatTwmXinos {
		pos.keyPatPos[k] = -1
	}

	for i, s := range tokens {
		if strings.HasSuffix(s, "_gp") {
			pos.posStart = i + 1
			break
		}

		switch s {
		case keyPatTwmXinos["MRBTSID"]:
			pos.keyPatPos["MRBTSID"] = i
		case keyPatTwmXinos["NRBTSID"]:
			pos.keyPatPos["NRBTSID"] = i
		case keyPatTwmXinos["NRCELLID"]:
			pos.keyPatPos["NRCELLID"] = i
		case keyPatTwmXinos["NRDUID"]:
			pos.keyPatPos["NRDUID"] = i
		case keyPatTwmXinos["NRCUUPID"]:
			pos.keyPatPos["NRCUUPID"] = i
		case keyPatTwmXinos["MCC"]:
			pos.keyPatPos["MCC"] = i
		case keyPatTwmXinos["MNC"]:
			pos.keyPatPos["MNC"] = i
		case keyPatTwmXinos["SST"]:
			pos.keyPatPos["SST"] = i
		case keyPatTwmXinos["SD"]:
			pos.keyPatPos["SD"] = i
		case keyPatTwmXinos["NRCI"]:
			pos.keyPatPos["NRCI"] = i
		case keyPatTwmXinos["ECI"]:
			pos.keyPatPos["ECI"] = i
		case keyPatTwmXinos["TS"]:
			pos.keyPatPos["TS"] = i
		}
	}

	return pos
}

func (p *PmParser) writeLog(level zapcore.Level, s string) {
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
