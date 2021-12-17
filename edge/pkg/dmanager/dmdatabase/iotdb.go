package dmdatabase

import (
	"github.com/apache/iotdb-client-go/rpc"
	"github.com/sirupsen/logrus"
	"log"
	"time"

	client "github.com/apache/iotdb-client-go/client"
)

type Record struct {
	DeviceID     	string
	Measurements 	[][]string
	Values       	[][]interface{}
	DataTypes   	[][]client.TSDataType
	Timestamp    	[]int64
	Sorted			bool
}

type OneRecord struct {
	DeviceID		string
	Measurements	[]string
	DataTypes		[]client.TSDataType
	Values			[]interface{}
	Timestamp		int64
}

func IotDBSessionInit() (*client.Session, error) {
	config := &client.Config{
		Host:     "192.168.0.3",
		Port:     "6667",
		UserName: "root",
		Password: "root",
	}
	session := client.NewSession(config)
	err := session.Open(false, 0)
	if err != nil {
		log.Fatal(err)
	}
	return session, err
}

func checkError(status *rpc.TSStatus, err error) error{
	if err != nil {
		logrus.Errorf("RPC error : ", err)
		return err
	}
	if status != nil {
		if err = client.VerifySuccess(status); err != nil{
			logrus.Errorf("RPC verify err :", err)
			return err
		}
	}
	return nil
}

func InsertDeviceRecord(session client.Session, DRecord Record) error {
	rStatus, err := session.InsertRecordsOfOneDevice(
		DRecord.DeviceID,
		DRecord.Timestamp,
		DRecord.Measurements,
		DRecord.DataTypes,
		DRecord.Values,
		DRecord.Sorted)
	err = checkError(rStatus, err)
	if err != nil {
		logrus.Errorf("Insert Device Record Error: ", err)
	}
	return err
}

func InsertOneRecord(session client.Session, oneRecord OneRecord) error {
	rStatus, err := session.InsertRecord(
		oneRecord.DeviceID,
		oneRecord.Measurements,
		oneRecord.DataTypes,
		oneRecord.Values,
		oneRecord.Timestamp)
	err = checkError(rStatus, err)
	if err != nil {
		logrus.Errorf("Insert One Record Error: ", err)
	}
	return err
}

func CreateOneRecord(DeviceID string, PropertyName string, DataType string, Data interface{}) OneRecord{
	var newDataType client.TSDataType
	switch DataType {
	case "int":
		newDataType = client.INT64
	case "double":
		newDataType = client.DOUBLE
	case "string":
		newDataType = client.TEXT
	case "float":
		newDataType = client.FLOAT
	case "bool":
		newDataType = client.BOOLEAN
	default:
		newDataType = client.UNKNOW
	}
	newRecord := OneRecord{
		DeviceID:     DeviceID,
		Measurements: []string{PropertyName},
		DataTypes:    []client.TSDataType{newDataType},
		Values:       []interface{}{Data},
		Timestamp:    time.Now().UTC().UnixNano() / 1000000,
	}
	return newRecord
}