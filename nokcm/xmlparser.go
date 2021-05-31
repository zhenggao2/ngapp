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
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"strings"
)

type XmlParser struct {
	log *zap.Logger
	debug bool
}

func (p *XmlParser) Init(log *zap.Logger, debug bool) {
	p.log = log
	p.debug = debug

	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Initializing XML parser..."))

}

func (p *XmlParser) Parse(xml, txml string) {
	switch strings.ToLower(txml) {
	case "scfc":
		p.ParseScfc(xml)
	case "vendor":
		p.ParseVendor(xml)
	case "freqhist":
		p.ParseFreqHist(xml)
	}
}

func (p *XmlParser) ParseScfc(xml string) {
	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Parsing SCFC...[%s]", xml))

}

func (p *XmlParser) ParseVendor(xml string) {
	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Parsing Vendor...[%s]", xml))

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
