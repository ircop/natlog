# RDP/EcoNat syslog parser and writer to ClickHouse


## Usage:

```bash
/path/to/natlog -c /etc/natlog.toml
```


-c: path to config file in toml format, as follows:

```
[listener]
listen-ip = "0.0.0.0"		# Which ip/port to listen for syslog messages
listen-port = 5555
workers = 32				# Workers count (they parse syslog messages)

# As you may know, data should not be written into clickhouse per row, but should be written with large banches:
[ch]
max-count       = 100000    # Maximum collected records to write in one banch
max-interval    = 30        # Maximum seconds between writes
connection-string = "tcp://127.0.0.1:9000?database=natlog&compress=true&debug=false"
```


After running, we can see something like this:

```
./natlog -c /etc/natlog.toml

{"level":"info","ts":1552476479.295355,"caller":"natlog/natlog.go:35","msg":"Starting syslog collector"}
{"level":"info","ts":1552476479.2960026,"caller":"chwriter/chwriter.go:47","msg":"resetTimer()"}
{"level":"info","ts":1552476495.5292575,"caller":"chwriter/chwriter.go:84","msg":"writing to db","records":100001}
{"level":"info","ts":1552476495.8284068,"caller":"chwriter/chwriter.go:47","msg":"resetTimer()"}
{"level":"info","ts":1552476511.3948112,"caller":"chwriter/chwriter.go:84","msg":"writing to db","records":100001}
{"level":"info","ts":1552476511.6750898,"caller":"chwriter/chwriter.go:47","msg":"resetTimer()"}
```



## RDP requirements

Currently this tool parse syslog messages with following EcoNat settings:

```
use_hex_format off
log_on_release on
log_individual_conn on
strip_tags off
pack_msgs on
log_format syslog
```




## Wiewing results:

My ip is '10.20.30.40' and i just executed ping to '1.1.1.1':

```
$ clickhouse-client

USE natlog

select date, time, IPv4NumToString(dst_ip), IPv4NumToString(nat_ip), IPv4NumToString(local_ip), \
dst_port, nat_port, local_port, proto, type \
from connections where dst_ip=IPv4StringToNum('1.1.1.1') AND local_ip=IPv4StringToNum('10.20.30.40')


┌───────date─┬────────────────time─┬─IPv4NumToString(dst_ip)─┬─IPv4NumToString(nat_ip)─┬─IPv4NumToString(local_ip)─┬─dst_port─┬─nat_port─┬─local_port─┬─proto─┬─type─┐
│ 2019-03-13 │ 2019-03-13 14:50:05 │ 1.1.1.1                 │ 101.102.103.104         │ 10.20.30.40               │        1 │        1 │          1 │ ICMP  │ E    │
└────────────┴─────────────────────┴─────────────────────────┴─────────────────────────┴───────────────────────────┴──────────┴──────────┴────────────┴───────┴──────┘

```


