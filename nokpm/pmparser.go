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
	cmap "github.com/orcaman/concurrent-map"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
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
	"NRCELL_PLMN_NRRELE" : "NRBTSID_NRCELLID_MCC_MNC_ECI_DMCC_DMNC_TS",
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
	"NGCFB" : "NRCELL",
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
	"NREMO" : "NRCELL_PLMN_NRRELE", // FIXME, new in 5G21A
	"NPSL" : "NRBTS_PLMN_SLICE", // new in 5G21A
	"NRMOP" : "NRCELL", // new in 5G21B
	"NRPFW" : "NRBTS", // new in 5G21B
	"NCAD1" : "NRDU", // new in 5G21B
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
	"M55157" : "NRMOP",
	"M55337" : "NRPFW",
	"M55348" : "NCAD1",
	"M55604" : "RURWS",
	"M55605" : "TRRW",
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
	"NRMOP" : "M55157",
	"NRPFW" : "M55337",
	"NCAD1" : "M55348",
	"RURWS" : "M55604",
	"TRRW" : "M55605",
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
	"DMCC" : "DMCC",
	"DMNC" : "DMNC",
	"TS" : "TIME",
}

// keyPatRawPm is map between token of the default key pattern and the field of Raw PM.
var keyPatRawPm = map[string]string {
	"NRBTSID" : "NRBTS",
	"NRCELLID" : "NRCELL",
	"NRDUID" : "NRDU",
	"NRCUUPID" : "NRCUUP",
	"MCC" : "MCC",
	"MNC" : "MNC",
	"SST" : "SST",
	"SD" : "SD",
	"NRCI" : "NRCI",
	"ECI" : "ECI",
	"DMCC" : "DMCC",
	"DMNC" : "DMNC",
	"TS" : "TS", // format: xxx/TS-startTime.interval
}

type PmParser struct {
	log   *zap.Logger
	op    string
	db    string
	debug bool
	rawDat cmap.ConcurrentMap
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
	p.rawDat = cmap.New()

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

func (p *PmParser) ArchiveRawPm() {
	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Archiving raw PMs..."))
	for _, counter := range p.rawDat.Keys() {
		ofn := filepath.Join(p.db, fmt.Sprintf("%s.gz", counter))
		fout, err := os.OpenFile(ofn, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0664)
		if err != nil {
			p.writeLog(zapcore.ErrorLevel, fmt.Sprintf("Fail to open file: %s", ofn))
			return
		}

		gw := gzip.NewWriter(fout)
		m, _ := p.rawDat.Get(counter)
		for _, key := range m.(cmap.ConcurrentMap).Keys() {
			val, _ := m.(cmap.ConcurrentMap).Get(key)
			s := fmt.Sprintf("%v,%v\n", key, val)
			gw.Write([]byte(s))
		}
		gw.Close()
		fout.Close()
	}
}

func (p *PmParser) ParseRawPmXml(xml string) {
	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("  Parsing raw PM...[%s]", xml))

	doc := etree.NewDocument()
	if err := doc.ReadFromFile(xml); err != nil {
		p.writeLog(zapcore.ErrorLevel, fmt.Sprintf("Fail to open file: %s", xml))
		return
	}

	omes := doc.SelectElement("OMeS")
	if omes == nil {
		p.writeLog(zapcore.DebugLevel, fmt.Sprintf("No OMeS element, xml=[%s]", xml))
		return
	}
	//fmt.Printf("[%s]: ns=%v, path=%v, index=%v, tag=%v, attr=%v\n", filepath.Base(xml), omes.NamespaceURI(), omes.GetPath(), omes.Index(), omes.Tag, omes.Attr)

	for _, pmSetup := range omes.FindElements("PMSetup") {
		//fmt.Printf("[%s]: ns=%v, path=%v, index=%v, tag=%v, attr=%v\n", filepath.Base(xml), pmSetup.NamespaceURI(), pmSetup.GetPath(), pmSetup.Index(), pmSetup.Tag, pmSetup.Attr)
		startTime := pmSetup.SelectAttrValue("startTime", "")
		t, _ := time.Parse("2006-01-02T15:04:05.000-07:00:00", startTime)
		startTime = t.Format("20060102150405")
		interval := pmSetup.SelectAttrValue("interval", "")

		for _, pmMoResult := range pmSetup.FindElements("PMMOResult") {
			moList := make([]string, 0)
			for _, mo := range pmMoResult.FindElements("MO") {
				dim := mo.SelectAttrValue("dimension", "")
				subDn := mo.FindElement("DN")
				if dim == "network_element" {
					if len(strings.Split(subDn.Text(), "/")) > 2 {
						moList = append(moList, strings.SplitN(subDn.Text(), "/", 3)[2]) // remove PLMN-PLMN/MRBTS-xxx
					} else {
						// <PMMOResult><MO dimension="network_element"><DN>PLMN-PLMN/MRBTS-1619377</DN></MO><NE-WBTS_1.0 measurementType="SBTS_Energy_Consumption"><M40002C1>17472</M40002C1><M40002C2>22740</M40002C2><M40002C0>5268</M40002C0></NE-WBTS_1.0></PMMOResult>
						moList = append(moList, strings.SplitN(subDn.Text(), "/", 3)[1]) // remove PLMN-PLMN
					}
				} else {
					moList = append(moList, strings.SplitN(subDn.Text(), "/", 2)[1]) // remove PLMN-PLMN
				}
			}

			// append 'TS'
			moList = append(moList, fmt.Sprintf("TS-%s.%s", startTime, interval))
			dn := strings.Join(moList, "/")

			// extract key
			keyTokens := make([]string, 0)
			t := strings.Split(dn, "/")
			for _, t2 := range t {
				keyTokens = append(keyTokens, strings.Split(t2, "-")[1])
			}
			key := strings.Join(keyTokens, "_")

			for _, neWbts := range pmMoResult.FindElements("NE-WBTS_1.0") {
				// measType := neWbts.SelectAttrValue("measurementType", "")
				for _, pm := range neWbts.ChildElements() {
					p.rawDat.SetIfAbsent(pm.Tag, cmap.New())
					m, _ := p.rawDat.Get(pm.Tag)
					m.(cmap.ConcurrentMap).SetIfAbsent(key, pm.Text())
					p.rawDat.Set(pm.Tag, m)
				}
			}
		}
	}

	/*
	for counter := range data {
		ofn := filepath.Join(p.db, fmt.Sprintf("%s.gz", counter))
		fout, err := os.OpenFile(ofn, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0664)
		if err != nil {
			p.writeLog(zapcore.ErrorLevel, fmt.Sprintf("Fail to open file: %s", ofn))
			return
		}

		gw := gzip.NewWriter(fout)
		gw.Write(data[counter].Bytes())
		gw.Close()
		fout.Close()
	}
	 */
}

func (p *PmParser) ParseSqlQueryCsv(csv string) {
	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Parsing csv PM(operator:%s)...[%s]", p.op, filepath.Base(csv)))

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

	// For TWM XINOS M55145(NRANS), the aggregation is NRBTS_PLMN
	var agg string
	if p.op == "twm" && measType == "NRANS" {
		agg = "NRBTS_PLMN"
	} else {
		agg = aggPerMeasType[measType]
	}

	buffPool := sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	data := make(map[string]*bytes.Buffer)
	maxLineToFlush := 20000
	nline := 0
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		// remove leading and tailing spaces
		line = strings.TrimSpace(line)
		tokens := strings.Split(line, ",")
		keyPatTokens := strings.Split(keyPerAgg[agg], "_")
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
				//data[counter] = &bytes.Buffer{}
				data[counter] = buffPool.Get().(*bytes.Buffer)
				data[counter].Reset()
			}
			data[counter].WriteString(key + "," + val + "\n")
		}

		nline += 1
		if nline >= maxLineToFlush {
			//fmt.Printf("nline=%d, writing to gz...[%s]\n", nline, filepath.Base(csv))
			for counter := range data {
				ofn := filepath.Join(p.db, fmt.Sprintf("%s.gz", counter))
				fout, err := os.OpenFile(ofn, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0664)
				if err != nil {
					p.writeLog(zapcore.ErrorLevel, fmt.Sprintf("Fail to open file: %s", ofn))
					break
				}

				gw := gzip.NewWriter(fout)
				gw.Write(data[counter].Bytes())
				gw.Close()

				data[counter].Reset()
				buffPool.Put(data[counter])

				fout.Close()
			}
			nline = 0
		}
	}

	fin.Close()

	fmt.Printf("writing last piece to gz...[%s]\n", filepath.Base(csv))
	for counter := range data {
		ofn := filepath.Join(p.db, fmt.Sprintf("%s.gz", counter))
		fout, err := os.OpenFile(ofn, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0664)
		if err != nil {
			p.writeLog(zapcore.ErrorLevel, fmt.Sprintf("Fail to open file: %s", ofn))
			break
		}

		gw := gzip.NewWriter(fout)
		gw.Write(data[counter].Bytes())
		buffPool.Put(data[counter])
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
		case keyPatTwmXinos["DMCC"]:
			pos.keyPatPos["DMCC"] = i
		case keyPatTwmXinos["DMNC"]:
			pos.keyPatPos["DMNC"] = i
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
