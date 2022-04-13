package main

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

func TestCRC(t *testing.T) {
	for i, tt := range []struct {
		data    string
		wantCRC string
	}{
		{data: "123456789", wantCRC: "BB3D"},
		{data: "1234567890", wantCRC: "C57A"},
		{data: "/ELL5\x5c253833635_A\r\n\r\n0-0:1.0.0(210217184019W)\r\n1-0:1.8.0(00006678.394*kWh)\r\n1-0:2.8.0(00000000.000*kWh)\r\n1-0:3.8.0(00000021.988*kvarh)\r\n1-0:4.8.0(00001020.971*kvarh)\r\n1-0:1.7.0(0001.727*kW)\r\n1-0:2.7.0(0000.000*kW)\r\n1-0:3.7.0(0000.000*kvar)\r\n1-0:4.7.0(0000.309*kvar)\r\n1-0:21.7.0(0001.023*kW)\r\n1-0:41.7.0(0000.350*kW)\r\n1-0:61.7.0(0000.353*kW)\r\n1-0:22.7.0(0000.000*kW)\r\n1-0:42.7.0(0000.000*kW)\r\n1-0:62.7.0(0000.000*kW)\r\n1-0:23.7.0(0000.000*kvar)\r\n1-0:43.7.0(0000.000*kvar)\r\n1-0:63.7.0(0000.000*kvar)\r\n1-0:24.7.0(0000.009*kvar)\r\n1-0:44.7.0(0000.161*kvar)\r\n1-0:64.7.0(0000.138*kvar)\r\n1-0:32.7.0(240.3*V)\r\n1-0:52.7.0(240.1*V)\r\n1-0:72.7.0(241.3*V)\r\n1-0:31.7.0(004.2*A)\r\n1-0:51.7.0(001.6*A)\r\n1-0:71.7.0(001.7*A)\r\n!", wantCRC: "7945"},
		{data: "/LGF5E360\r\n\r\n0-0:1.0.0(220330220600W)\r\n1-0:1.8.0(00005895.962*kWh)\r\n1-0:2.8.0(00001044.196*kWh)\r\n1-0:3.8.0(00001783.044*kVArh)\r\n1-0:4.8.0(00000454.226*kVArh)\r\n1-0:1.7.0(0000.699*kW)\r\n1-0:2.7.0(0000.000*kW)\r\n1-0:3.7.0(0000.000*kVAr)\r\n1-0:4.7.0(0000.028*kVAr)\r\n1-0:21.7.0(0000.121*kW)\r\n1-0:22.7.0(0000.000*kW)\r\n1-0:41.7.0(0000.152*kW)\r\n1-0:42.7.0(0000.000*kW)\r\n1-0:61.7.0(0000.425*kW)\r\n1-0:62.7.0(0000.000*kW)\r\n1-0:23.7.0(0000.000*kVAr)\r\n1-0:24.7.0(0000.045*kVAr)\r\n1-0:43.7.0(0000.000*kVAr)\r\n1-0:44.7.0(0000.081*kVAr)\r\n1-0:63.7.0(0000.098*kVAr)\r\n1-0:64.7.0(0000.000*kVAr)\r\n1-0:32.7.0(228.4*V)\r\n1-0:52.7.0(229.1*V)\r\n1-0:72.7.0(228.3*V)\r\n1-0:31.7.0(000.5*A)\r\n1-0:51.7.0(000.7*A)\r\n1-0:71.7.0(001.9*A)\r\n!", wantCRC: "830C"},
	} {
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			got := crc16Sum(crc16(0, []byte(tt.data)))
			if got != tt.wantCRC {
				t.Errorf("ERROR: got CRC16(1): %s, want: %s", got, tt.wantCRC)
			}
		})
	}
}

func TestDecoder_ReadBlock(t *testing.T) {
	for i, tt := range []struct {
		data     string
		wantID   string
		wantData SmartEnergyMeterData
	}{
		{
			data:   "/ELL5\x5c253833635_A\r\n\r\n0-0:1.0.0(210217184019W)\r\n1-0:1.8.0(00006678.394*kWh)\r\n1-0:2.8.0(00000000.000*kWh)\r\n1-0:3.8.0(00000021.988*kvarh)\r\n1-0:4.8.0(00001020.971*kvarh)\r\n1-0:1.7.0(0001.727*kW)\r\n1-0:2.7.0(0000.000*kW)\r\n1-0:3.7.0(0000.000*kvar)\r\n1-0:4.7.0(0000.309*kvar)\r\n1-0:21.7.0(0001.023*kW)\r\n1-0:41.7.0(0000.350*kW)\r\n1-0:61.7.0(0000.353*kW)\r\n1-0:22.7.0(0000.000*kW)\r\n1-0:42.7.0(0000.000*kW)\r\n1-0:62.7.0(0000.000*kW)\r\n1-0:23.7.0(0000.000*kvar)\r\n1-0:43.7.0(0000.000*kvar)\r\n1-0:63.7.0(0000.000*kvar)\r\n1-0:24.7.0(0000.009*kvar)\r\n1-0:44.7.0(0000.161*kvar)\r\n1-0:64.7.0(0000.138*kvar)\r\n1-0:32.7.0(240.3*V)\r\n1-0:52.7.0(240.1*V)\r\n1-0:72.7.0(241.3*V)\r\n1-0:31.7.0(004.2*A)\r\n1-0:51.7.0(001.6*A)\r\n1-0:71.7.0(001.7*A)\r\n!7945\r\n",
			wantID: "ELL5\\253833635_A",
			wantData: SmartEnergyMeterData{
				ID: "ELL5\\253833635_A",
				Values: []ObisData{
					{OBIS: "1-0:1.8.0", Value: 6678.394, Unit: "kWh"},
					{OBIS: "1-0:2.8.0", Value: 0, Unit: "kWh"},
					{OBIS: "1-0:3.8.0", Value: 21.988, Unit: "kvarh"},
					{OBIS: "1-0:4.8.0", Value: 1020.971, Unit: "kvarh"},
					{OBIS: "1-0:1.7.0", Value: 1.727, Unit: "kW"},
					{OBIS: "1-0:2.7.0", Value: 0, Unit: "kW"},
					{OBIS: "1-0:3.7.0", Value: 0, Unit: "kvar"},
					{OBIS: "1-0:4.7.0", Value: 0.309, Unit: "kvar"},
					{OBIS: "1-0:21.7.0", Value: 1.023, Unit: "kW"},
					{OBIS: "1-0:41.7.0", Value: 0.350, Unit: "kW"},
					{OBIS: "1-0:61.7.0", Value: 0.353, Unit: "kW"},
					{OBIS: "1-0:22.7.0", Value: 0, Unit: "kW"},
					{OBIS: "1-0:42.7.0", Value: 0, Unit: "kW"},
					{OBIS: "1-0:62.7.0", Value: 0, Unit: "kW"},
					{OBIS: "1-0:23.7.0", Value: 0, Unit: "kvar"},
					{OBIS: "1-0:43.7.0", Value: 0, Unit: "kvar"},
					{OBIS: "1-0:63.7.0", Value: 0, Unit: "kvar"},
					{OBIS: "1-0:24.7.0", Value: 0.009, Unit: "kvar"},
					{OBIS: "1-0:44.7.0", Value: 0.161, Unit: "kvar"},
					{OBIS: "1-0:64.7.0", Value: 0.138, Unit: "kvar"},
					{OBIS: "1-0:32.7.0", Value: 240.3, Unit: "V"},
					{OBIS: "1-0:52.7.0", Value: 240.1, Unit: "V"},
					{OBIS: "1-0:72.7.0", Value: 241.3, Unit: "V"},
					{OBIS: "1-0:31.7.0", Value: 4.2, Unit: "A"},
					{OBIS: "1-0:51.7.0", Value: 1.6, Unit: "A"},
					{OBIS: "1-0:71.7.0", Value: 1.7, Unit: "A"},
				},
			},
		},
		// TODO: Add more test cases.
	} {
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			d := NewDecoder(bytes.NewBufferString(tt.data))
			got, err := d.ReadBlock()
			if err != nil {
				t.Fatalf("ERROR: got error: %v", err)
			}
			if got == nil {
				t.Fatal("ERROR: got == nil")
			}
			if got.ID != tt.wantID {
				t.Errorf("ERROR: got ident: %s, want %s", got.ID, tt.wantID)
			}
			if !reflect.DeepEqual(*got, tt.wantData) {
				t.Errorf("ERROR: got: %v, want: %v", got, tt.wantData)
			}
		})
	}
}
