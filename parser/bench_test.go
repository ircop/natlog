package parser

import (
	"regexp"
	"testing"
)

func BenchmarkParseFields(b *testing.B) {
	str := `2019-03-12T22:01:32+03:00 nat-100.ip-ho |  151.101.246.002:00443 F 091.245.129.236:01049 E 010.010.013.140:50671   TCP 005.189.157.090:02517 A 091.245.129.236:01025 I 010.010.013.140:14567   UDP`
	re := regexp.MustCompile(`(?msi:\s+(?P<dstip>\d+.\d+.\d+.\d+):(?P<dstport>\d+)\s+(?P<action>(A|F|E|I))\s+(?P<natip>[^:]+):(?P<natport>\d+))\s+(?P<type>(A|F|E|I))\s+(?P<localip>[^:]+):(?P<localport>\d+)\s+(?P<proto>[a-zA-Z]+)`)

	ts, _ := getTimeWithSplit(&str)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := parseFields(&str, re, ts); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetTimeRegex(b *testing.B) {
	str := "2019-03-12T22:01:32+03:00 nat-100.ip-ho |  151.101.246.002:00443 F 091.245.129.236:01049 E 010.010.013.140:50671   TCP 005.189.157.090:02517 A 091.245.129.236:01025 I 010.010.013.140:14567   UDP"
	re := regexp.MustCompile(`(?msi:^[^\s]+)`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := getTimeWithRegex(&str, re); err != nil {
			b.Fatal(err)
		}
	}
}


func BenchmarkGetTimeWithSplit(b *testing.B) {
	str := "2019-03-12T22:01:32+03:00 nat-100.ip-ho |  151.101.246.002:00443 F 091.245.129.236:01049 E 010.010.013.140:50671   TCP 005.189.157.090:02517 A 091.245.129.236:01025 I 010.010.013.140:14567   UDP"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := getTimeWithSplit(&str); err != nil {
			b.Fatal(err)
		}
	}
}
