package parser

import (
	"errors"
	"go.uber.org/zap"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//type ErrTimeNotFound error
var ErrTimeNotFound = errors.New("Time not found in string")
var ErrParsingError	= errors.New("Failed to parse record")

var timeRe *regexp.Regexp
var strRe *regexp.Regexp
var dataChan chan *NatRecord

type PROTO	int32
const (
	UNKNOWN PROTO = 0
	TCP		PROTO = 1
	UDP		PROTO = 2
	ICMP	PROTO = 3
)
func (p PROTO) String() string {
	return []string{"UNKNOWN", "TCP", "UDP", "ICMP"}[p]
}

type NatRecord struct {
	Time		*time.Time
	DstIP		net.IP
	DstPort		int64
	LocalIP		net.IP
	LocalPort	int64
	NatIP		net.IP
	NatPort		int64
	Proto		PROTO
	Action		string
	Type		string
}

func Init(ch chan *NatRecord) {
	timeRe = regexp.MustCompile(`(?msi:^[^\s]+)`)
	strRe = regexp.MustCompile(`(?msi:\s+(?P<dstip>\d+.\d+.\d+.\d+):(?P<dstport>\d+)\s+(?P<action>(A|F|E|I))\s+(?P<natip>[^:]+):(?P<natport>\d+))\s+(?P<type>(A|F|E|I))\s+(?P<localip>[^:]+):(?P<localport>\d+)\s+(?P<proto>[a-zA-Z]+)`)
	dataChan = ch
}

func ParseMessage(msg string) {
	//fmt.Printf("msg: '%s'\n\n", msg)
	ts, err := getTimeWithSplit(&msg)
	if err != nil {
		zap.L().Error("failed to parse time", zap.Error(err))
		return
	}

	records, err := parseFields(&msg, strRe, ts)
	if err != nil {
		zap.L().Error("Failed to parse fields", zap.Error(err))
		return
	}
	//fmt.Printf("results: %+#v\n", results)

	for i := range records {

		dataChan <- records[i]

		/*zap.L().Info("nat-record",
			zap.Time("ts", *records[i].Time),
			zap.String("dst-ip", records[i].DstIP.String()),
			zap.Int64("dst-port", records[i].DstPort),
			zap.String("nat-ip", records[i].NatIP.String()),
			zap.Int64("nat-port", records[i].NatPort),
			zap.String("local-ip", records[i].LocalIP.String()),
			zap.Int64("local-port", records[i].LocalPort),
			zap.Int32("proto", int32(records[i].Proto)),
			)*/
	}
}

func parseFields(msg *string, re *regexp.Regexp, timestamp *time.Time) ([]*NatRecord, error) {
	results := make([]map[string]string, 0)
	matches := re.FindAllStringSubmatch(*msg, -1)
	names := re.SubexpNames()

	for i := range matches {
		curMap := make(map[string]string)
		for j, name := range names {
			if name != "" {
				curMap[name] = matches[i][j]
			}
		}
		results = append(results, curMap)
	}

	records := make([]*NatRecord, 0)
	for i := range results {
		//fmt.Printf("res: %+#v\n", results[i])
		r := NatRecord{
			Time:		timestamp,
			DstIP:		net.ParseIP(results[i]["dstip"]),
			LocalIP:	net.ParseIP(results[i]["localip"]),
			NatIP:		net.ParseIP(results[i]["natip"]),
			Action:		results[i]["action"],
			Type:		results[i]["type"],
			Proto:		UNKNOWN,
		}

		if r.DstIP == nil || r.LocalIP == nil || r.NatIP == nil || len(r.Action) != 1 || len(r.Type) != 1 {
			return records, ErrParsingError
		}
		if dport, err := strconv.ParseInt(results[i]["dstport"], 10, 64); err != nil {
			return records, ErrParsingError
		}  else {
			r.DstPort = dport
		}
		if nport, err := strconv.ParseInt(results[i]["natport"], 10, 64); err != nil {
			return records, ErrParsingError
		} else {
			r.NatPort = nport
		}
		if lport, err := strconv.ParseInt(results[i]["localport"], 10, 64); err != nil {
			return records, ErrParsingError
		} else {
			r.LocalPort = lport
		}
		switch results[i]["proto"] {
		case "TCP":
			r.Proto = TCP
		case "UDP":
			r.Proto = UDP
		case "ICMP", "ICM":
			r.Proto = ICMP
		}

		records = append(records, &r)
	}

	return records, nil
}

func getTimeWithRegex(msg *string, re *regexp.Regexp) (*time.Time, error) {
	str := re.FindString(*msg)
	if str == "" {
		return nil, ErrTimeNotFound
	}

	ts, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return nil, err
	}

	return &ts, nil
}

func getTimeWithSplit(msg *string) (*time.Time, error) {
	spl := strings.Split(*msg, " ")
	if len(spl) < 2 {
		return nil, ErrTimeNotFound
	}

	ts, err := time.Parse(time.RFC3339, spl[0])
	if err != nil {
		return nil, err
	}

	return &ts, nil
}

/*
func FirstParseMessage(msg string) {
	spl := strings.Split(msg, "|")
	if len(spl) != 2 {
		return
	}

	ts, err := getTimestamp(&spl[0])
	if err != nil {
		zap.L().Error("Failed to parse timestamp", zap.Error(err))
		return
	}
	zap.L().Info("parsed", zap.String("ts", ts.String()), zap.Int64("unix", ts.Unix()))
	zap.L().Info("ts parsed, go next", zap.String("str", spl[1]))
}

func getTimestamp(part *string) (*time.Time, error) {
	s := strings.Split(*part, " ")
	if len(s) < 2 {
		return nil, errors.New("Wrong string format: cannot split by space")
	}

	ts, err := time.Parse(time.RFC3339, s[0])
	if err != nil {
		return nil, err
	}
	return &ts, nil
}
*/

/*

create table natlog.connections ( \
 date Date, \
 time DateTime, \
 dst_ip UInt32, \
 nat_ip UInt32, \
 local_ip UInt32, \
 dst_port UInt16, \
 nat_port UInt16, \
 local_port UInt16, \
 proto Enum8('UNKNOWN' = 0, 'TCP' = 1, 'UDP' = 2, 'ICMP' = 3), \
 type FixedString(2) \
) engine=MergeTree(date, (dst_ip, nat_ip, local_ip), 8192)

 */