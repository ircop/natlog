package chwriter

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

import (
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/ircop/natlog/parser"
	_ "github.com/kshvakov/clickhouse"
	"go.uber.org/zap"
	"sync"
	"time"
)

var ErrWrongInterval = errors.New("Wrong write interval")
var ErrWrongCount = errors.New("Wrong write max count")

type Chwriter struct {
	ConnectString		string
	WriteMaxInterval	int64			// make write every N seconds
	WriteMaxCount		int64			// or when N records collected
	Conn				*sql.DB

	DataMX				sync.Mutex
	Data				[]*parser.NatRecord
	DataTimer			*time.Timer
}

var Writer Chwriter

func (w *Chwriter) resetTimer() {
	zap.L().Info("resetTimer()")

	w.DataMX.Lock()
	defer w.DataMX.Unlock()

	if w.DataTimer != nil {
		w.DataTimer.Stop()
		w.DataTimer = nil
	}

	w.DataTimer = time.AfterFunc(time.Duration(w.WriteMaxInterval) * time.Second, func () {
		w.DataMX.Lock()
		zap.L().Info("Writing on-timer", zap.Int("count", len(w.Data)))

		if len(w.Data) > 0 {
			go w.write(w.Data)
			w.Data = make([]*parser.NatRecord, 0)
			w.DataMX.Unlock()
		} else {
			w.DataMX.Unlock()
			w.resetTimer()
		}
	})
}

func (w *Chwriter) write(data []*parser.NatRecord) {
	defer func() {
		if r := recover(); r != nil {
			zap.L().Error("panic in write()", zap.String("panic", fmt.Sprintf("%v", r)))
		}
		w.resetTimer()
	}()

	if len(data) == 0 {
		return
	}

	zap.L().Info("writing to db", zap.Int("records", len(data)))

	tx, err := w.Conn.Begin()
	if err != nil {
		zap.L().Error("cannot begin transaction", zap.Error(err))
		return
	}

	stmt, err := tx.Prepare(`INSERT INTO connections(date, time, dst_ip, nat_ip, local_ip, dst_port, nat_port, local_port, proto, type) VALUES(?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		zap.L().Error("Cannot prepare statement", zap.Error(err))
		return
	}
	defer stmt.Close()


	for i := range data {
		if _, err = stmt.Exec(
				data[i].Time,
				data[i].Time,
				binary.BigEndian.Uint32(data[i].DstIP[12:16]),
				binary.BigEndian.Uint32(data[i].NatIP[12:16]),
				binary.BigEndian.Uint32(data[i].LocalIP[12:16]),
				data[i].DstPort,
				data[i].NatPort,
				data[i].LocalPort,
				data[i].Proto.String(),
				data[i].Type,
			); err != nil {
				zap.L().Error("Failed to exec statement", zap.Error(err))
			}
	}

	if err = tx.Commit(); err != nil {
		zap.L().Error("Failed to commit transaction", zap.Error(err))
	}
}

func (w *Chwriter) listen(ch chan *parser.NatRecord) {
	for {
		select {
		case nr := <- ch:
			w.DataMX.Lock()
			w.Data = append(w.Data, nr)
			if int64(len(w.Data)) > w.WriteMaxCount {
				go w.write(w.Data)
				w.Data = make([]*parser.NatRecord, 0)
			}
			w.DataMX.Unlock()
			break
		}
	}
}

func Init(cstring string, interval int64, cnt int64, dataChan chan *parser.NatRecord) error {
	if interval < 1 {
		return ErrWrongInterval
	}
	if cnt < 1 {
		return ErrWrongCount
	}

	Writer.ConnectString = cstring
	Writer.WriteMaxInterval = interval
	Writer.WriteMaxCount = cnt

	var err error
	Writer.Conn, err = sql.Open("clickhouse", Writer.ConnectString)
	if err != nil {
		return err
	}
	if err = Writer.Conn.Ping(); err != nil {
		return err
	}

	Writer.Data = make([]*parser.NatRecord, 0)

	Writer.resetTimer()
	go Writer.listen(dataChan)

	return nil
}
