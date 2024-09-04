package main

import (
	"flag"
	"strings"

	"github.com/curtisnewbie/miso/encoding"
	"github.com/curtisnewbie/miso/util"
	"golang.design/x/clipboard"
)

var (
	Input string
	Debug bool
)

func main() {
	flag.BoolVar(&Debug, "debug", false, "Debug")
	flag.StringVar(&Input, "input", "", "Input File")
	flag.Parse()

	var curl string
	if Input != "" {
		f, err := util.ReadFileAll(Input)
		util.Must(err)
		curl = util.UnsafeByt2Str(f)
	} else {
		err := clipboard.Init()
		util.DebugPrintlnf(Debug, "clipboard init")
		if err == nil {
			txt := clipboard.Read(clipboard.FmtText)
			if txt != nil {
				s := util.UnsafeByt2Str(txt)
				if strings.Contains(strings.ToLower(s), "curl") {
					curl = s
				}
			}
		}
	}

	if curl == "" {
		util.Printlnf("Missing curl command, either specify input file or copy the curl command to clipboard.")
		return
	}

	inst, ok := TryParseCurl(curl)
	if !ok {
		util.Printlnf("Failed to parse curl command")
		return
	}

	util.DebugPrintlnf(Debug, "%#v", inst)

	py := GenRequests(inst)
	print(py)
	println()
}

func GenRequests(inst Instruction) string {
	headers := "{}"
	data := "None"
	if len(inst.Headers) > 0 {
		headers, _ = encoding.SWriteJson(inst.Headers)
	}
	if inst.Payload != "" {
		data = inst.Payload
	}

	return util.NamedSprintf(`
import requests

data = ${data}
headers = ${headers}
res = requests.${method}(url='${url}', data=data, headers=headers)
print(res.status_code)
print(res.text)
	`,
		map[string]any{
			"method":  strings.ToLower(inst.Method),
			"url":     inst.Url,
			"data":    data,
			"headers": headers,
		})
}

type Instruction struct {
	Url     string
	Method  string
	Headers map[string]string
	Payload string
}

// TODO: improve this parser, it's now only useful for well-structured curl 'copied' from Chrome
func TryParseCurl(curl string) (inst Instruction, ok bool) {
	if util.IsBlankStr(curl) {
		return
	}
	inst.Headers = map[string]string{}
	if util.IsBlankStr(inst.Method) {
		inst.Method = "GET"
	}

	segments := curlSegments(curl)
	for i := range segments {
		seg := strings.TrimSpace(segments[i])

		if k, v, ok := parseCurlParam(seg, "-H"); ok { // header
			inst.Headers[k] = v
		} else if v, ok := parseData(seg, "-d"); ok { // body
			inst.Payload = v
		} else if v, ok := parseData(seg, "--data-raw"); ok { // body
			inst.Payload = v
		} else if _, v, ok := parseCurlParam(seg, "-X"); ok { // method
			inst.Method = v
		} else if v, ok := parseCurlDest(seg); ok { // destination
			inst.Url = v
		}
	}
	if inst.Method == "GET" && inst.Payload != "" {
		inst.Method = "POST"
	}

	util.DebugPrintlnf(Debug, "%+v", inst)
	ok = true
	return
}

func unquote(s string) string {
	s = strings.TrimSpace(s)
	v := []rune(s)
	if len(v) >= 2 && (v[0] == '\'' || v[0] == '"') {
		return string(v[1 : len(v)-1])
	}
	return strings.TrimSpace(string(v))
}

func parseData(seg string, prefix string) (string, bool) {
	if strings.HasPrefix(seg, prefix) {
		return string([]rune(seg)[len([]rune(prefix)):]), true
	}
	return "", false
}

func parseCurlParam(seg string, prefix string) (string, string, bool) {
	if strings.HasPrefix(seg, prefix) {
		seg = unquote(string([]rune(seg)[len([]rune(prefix)):]))
		util.DebugPrintlnf(Debug, "seg: %v", seg)
		tokens := strings.SplitN(seg, ":", 2)
		if len(tokens) > 1 { // k : value
			k := strings.TrimSpace(tokens[0])
			v := strings.TrimSpace(tokens[1])
			return k, v, true
		}
		if len(tokens) > 0 { // only value
			val := strings.TrimSpace(tokens[0])
			return "", val, true
		}
	}
	return "", "", false
}

func parseCurlDest(v string) (string, bool) {
	if j := strings.Index(v, "http"); j > -1 { // it may look like 'curl "http:...." or "http:..."'
		s := []rune(v)[j:]
		util.DebugPrintlnf(Debug, "(http) s: %v, j: %v", v, j)
		k := len(s) - 1
		if s[k] == '\'' || s[k] == '"' {
			quote := s[k]
			for s[k] == quote {
				k--
			}
		}
		s = s[:k+1]
		return string(s), true
	}
	return "", false
}

func curlSegments(curl string) []string {
	// TODO: should support curl that are not so well structured
	return strings.Split(curl, "\\")
}
