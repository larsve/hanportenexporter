package main

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"
)

type (
	DataWriter interface {
		Write(*SmartEnergyMeterData)
	}
	SerialBridgeClient struct {
		addr string
		conn net.Conn
		data DataWriter
	}
)

var (
	wg     sync.WaitGroup
	dialer net.Dialer
)

func init() {
	dialer.Timeout = 10 * time.Second
}

func startSerialBridgeClient(ctx context.Context, addr string, data DataWriter) {
	log.Printf("Starting serial bridge client: %s", addr)
	wg.Add(1)
	sbc := &SerialBridgeClient{addr: addr, data: data}
	go sbc.run(ctx)
}

func stopSerialBridgeClients() {
	log.Println("Stopping serial bridge clients...")
	wg.Wait()
	log.Println("Serial bridge clients stopped")
}

func (sbc *SerialBridgeClient) connect(ctx context.Context, reconnectChan chan<- struct{}) error {
	var err error
	log.Printf("Connecting to %s..", sbc.addr)
	sbc.conn, err = dialer.DialContext(ctx, "tcp", sbc.addr)
	if err != nil {
		log.Printf("Could not connect to %s, error: %v", sbc.addr, err)
		return err
	}

	log.Printf("Connected to %s", sbc.addr)

	obis := NewDecoder(sbc.conn)
	go func() {
		for {
			err = sbc.conn.SetReadDeadline(time.Now().Add(25 * time.Second))
			if err != nil {
				log.Printf("SetReadDeadline failed with error: %v (addr: %s)", err, sbc.addr)
			}
			blk, err := obis.ReadBlock()
			if err != nil {
				if errors.Is(err, ErrCRCError) {
					var me *MessageError
					if errors.As(err, &me) {
						log.Printf("CRC ERROR\nCalculated CRC: %s\nMessage:\n%s", me.MessageCRC, me.Message)
					}
					continue
				}
				if ctx.Err() == nil {
					log.Printf("Error while reading from %v, error: %v", sbc.addr, err)
					reconnectChan <- struct{}{}
				}
				return
			}
			data.Write(blk)
		}
	}()
	return nil
}

const (
	reconnectBackoffIntervalMax  = 15 * time.Minute
	reconnectBackoffIntervalBase = 10 * time.Second
)

func (sbc *SerialBridgeClient) run(ctx context.Context) {
	var stopErr error
	defer func() {
		defer wg.Done()
		if stopErr != nil {
			log.Printf("Stopped serial bridge client: %s due to error: %v", sbc.addr, stopErr)
		} else {
			log.Printf("Stopped serial bridge client: %s", sbc.addr)
		}
	}()

	var reconnectTime <-chan time.Time
	reconnectBackoffInterval := reconnectBackoffIntervalBase
	reconnect := make(chan struct{})

	setReconnectTime := func() {
		interval := reconnectBackoffInterval + time.Duration(rand.Intn(int(reconnectBackoffInterval/10)))
		reconnectBackoffInterval *= 2
		if reconnectBackoffInterval > reconnectBackoffIntervalMax {
			reconnectBackoffInterval = reconnectBackoffIntervalMax
		}
		log.Printf("Reconnecting to %s in %v", sbc.addr, interval)
		reconnectTime = time.After(interval)
	}

	connect := func() {
		reconnectTime = nil
		err := sbc.connect(ctx, reconnect)
		if err != nil {
			setReconnectTime()
			return
		}

		// Reset reconnectBackoffInterval since we were able to connect to server
		reconnectBackoffInterval = reconnectBackoffIntervalBase
	}

	connect()

	for {
		select {
		case <-ctx.Done():
			if sbc.conn != nil {
				sbc.conn.Close()
			}
			stopErr = ctx.Err()
			return
		case <-reconnect:
			if sbc.conn != nil {
				sbc.conn.Close()
			}
			setReconnectTime()
		case <-reconnectTime:
			connect()
		}
	}
}
