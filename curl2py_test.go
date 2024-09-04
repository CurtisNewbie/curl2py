package main

import "testing"

func TestTryParseCurl(t *testing.T) {
	s1 := `
curl 'http://localhost/ping' \
  -H 'Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7' \
  -H 'Accept-Language: en,zh-CN;q=0.9,zh;q=0.8' \
  -H 'Cache-Control: no-cache' \
  -H 'Connection: keep-alive' \
  -H 'Pragma: no-cache' \
  -H 'Sec-Fetch-Dest: document' \
  -H 'Sec-Fetch-Mode: navigate' \
  -H 'Sec-Fetch-Site: none' \
  -H 'Sec-Fetch-User: ?1' \
  -H 'Upgrade-Insecure-Requests: 1' \
  -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36' \
  -H 'sec-ch-ua: "Chromium";v="128", "Not;A=Brand";v="24", "Google Chrome";v="128"' \
  -H 'sec-ch-ua-mobile: ?0' \
  -H 'sec-ch-ua-platform: "macOS"'
	`
	inst, ok := TryParseCurlAst(s1)
	if !ok {
		t.FailNow()
	}
	t.Logf("inst: %#v", inst)
}

func TestTryParseCurl2(t *testing.T) {
	s1 := `
curl 'http://localhost/ping' \
  -H 'Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7' \
  -H 'Accept-Language: en,zh-CN;q=0.9,zh;q=0.8' \
  -H 'Cache-Control: no-cache' \
  -H 'Connection: keep-alive' \
  -H 'Pragma: no-cache' \
  -H 'Sec-Fetch-Dest: document' \
  -H 'Sec-Fetch-Mode: navigate' \
  -H 'Sec-Fetch-Site: none' \
  -H 'Sec-Fetch-User: ?1' \
  -H 'Upgrade-Insecure-Requests: 1' \
  -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36' \
  -H 'sec-ch-ua: "Chromium";v="128", "Not;A=Brand";v="24", "Google Chrome";v="128"' \
  -H 'sec-ch-ua-mobile: ?0' \
  -H 'sec-ch-ua-platform: "macOS"' \
  -d '{}'
	`
	inst, ok := TryParseCurlAst(s1)
	if !ok {
		t.FailNow()
	}
	if inst.Method != "POST" {
		t.FailNow()
	}
	t.Logf("inst: %#v", inst)
}

func TestTryParseCurlUsingParser(t *testing.T) {
	s1 := `
curl -X POST 'http://localhost/ping'
  -H 'Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7'
  -H 'Accept-Language: en,zh-CN;q=0.9,zh;q=0.8'
  -H 'Cache-Control: no-cache'
  -H 'Connection: keep-alive'
  -H 'Pragma: no-cache'
  -H 'Sec-Fetch-Dest: document'
  -H 'Sec-Fetch-Mode: navigate'
  -H 'Sec-Fetch-Site: none'
  -H 'Sec-Fetch-User: ?1'
  -H 'Upgrade-Insecure-Requests: 1'
  -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36'
  -H 'sec-ch-ua: "Chromium";v="128", "Not;A=Brand";v="24", "Google Chrome";v="128"'
  -H 'sec-ch-ua-mobile: ?0'
  -H 'sec-ch-ua-platform: "macOS"'
  -d '{ "name" : "abc" }'
	`
	inst, ok := TryParseCurlAst(s1)
	if !ok {
		t.FailNow()
	}
	if inst.Method != "POST" {
		t.FailNow()
	}
	t.Logf("inst.url: %v", inst.Url)
	t.Logf("inst.method: %v", inst.Method)
	for k, v := range inst.Headers {
		t.Logf("header: %v -> %v", k, v)
	}
	t.Logf("inst.payload: %v", inst.Payload)
}
