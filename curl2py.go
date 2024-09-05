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

	inst, ok := ParseCurl(curl)
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

func ParseCurl(curl string) (inst Instruction, ok bool) {
	if util.IsBlankStr(curl) {
		return
	}
	inst.Headers = map[string]string{}
	if util.IsBlankStr(inst.Method) {
		inst.Method = "GET"
	}

	p := NewCurlParser(curl)
	for p.HasNext() {
		tok := p.Next()
		util.DebugPrintlnf(Debug, "next tok: %v", tok)
		switch tok {
		case "-H":
			k, v, ok := util.SplitKV(unquote(p.Next()), ":")
			if ok {
				inst.Headers[k] = v
			}
		case "-X":
			inst.Method = unquote(p.Next())
		case "-d", "--data-raw":
			inst.Payload = p.Next()
		case "curl":
		default:
			util.DebugPrintlnf(Debug, "default tok: %v", tok)
			if tok != "" {
				inst.Url = unquote(tok)
			}
		}
	}
	if inst.Method == "GET" && inst.Payload != "" {
		inst.Method = "POST"
	}

	util.DebugPrintlnf(Debug, "inst: %+v", inst)
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

func NewCurlParser(curl string) *CurlParser {
	rc := []rune(curl)
	return &CurlParser{curl: curl, rcurl: rc, pos: 0, rlen: len(rc)}
}

type CurlParser struct {
	curl  string
	rcurl []rune
	rlen  int
	pos   int
}

func (c *CurlParser) HasNext() bool {
	return c.pos < len(c.rcurl)
}

func (c *CurlParser) inRange(n int) bool {
	return c.pos+n < len(c.rcurl)
}

func (c *CurlParser) peek(n int) rune {
	return c.rcurl[c.pos+n]
}

func (c *CurlParser) move(n int) {
	c.pos += n
}

func (c *CurlParser) parseCmdKey() string {
	i := 0
	for c.inRange(i) {
		switch c.peek(i) {
		case ' ', '\t', '\n':
			s := c.rcurl[c.pos : c.pos+i]
			c.move(i)
			return string(s)
		}
		i++
	}
	return ""
}

func (c *CurlParser) parseStr() string {
	stack := util.NewStack[rune](10)
	cur := c.peek(0)
	stack.Push(cur)
	i := 1
	escape := false
	for c.inRange(i) && !stack.Empty() {
		p := c.peek(i)
		switch p {
		case '\'', '"':
			if escape {
				escape = false
			} else {
				if p == cur {
					stack.Pop()
					cur, _ = stack.Peek()
				} else {
					stack.Push(p)
					cur = p
				}
			}
		case '\\':
			escape = !escape
		}
		i++
	}
	s := c.rcurl[c.pos : c.pos+i]
	c.move(i)
	vs := string(s)
	util.DebugPrintlnf(Debug, "parseStr, s: %v", vs)
	return vs
}

func (c *CurlParser) isSpace(n int) bool {
	return c.peek(n) == ' ' || c.peek(n) == '\n' || c.peek(n) == '\t' || c.peek(n) == '\\'
}

func (c *CurlParser) skipSpaces() {
	for c.HasNext() && c.isSpace(0) {
		c.move(1)
	}
}

func (c *CurlParser) parseWords() string {
	c.skipSpaces()
	i := 0
	for c.inRange(i) && !c.isSpace(i) {
		i++
	}
	s := c.rcurl[c.pos : c.pos+i]
	c.move(i)
	return string(s)
}

func (c *CurlParser) Next() (tok string) {
	c.skipSpaces()

	if !c.HasNext() {
		return tok
	}

	curr := c.peek(0)
	switch curr {
	case '-':
		return c.parseCmdKey()
	case '\'', '"':
		return c.parseStr()
	default:
		return c.parseWords()
	}
}
