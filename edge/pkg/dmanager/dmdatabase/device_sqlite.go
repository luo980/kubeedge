package dmdatabase

import (
	"github.com/kubeedge/kubeedge/edge/pkg/common/dbm"
	"github.com/sirupsen/logrus"
	"k8s.io/klog/v2"
)

//DeviceDelete the struct for deleting device
type DeviceDelete struct {
	DeviceID string
	Name     string
}

//DeviceAttrUpdate the struct for updating device attr
type DeviceAttrUpdate struct {
	DeviceID string
	Name     string
	Cols     map[string]interface{}
}

//Device the struct of device
type Device struct {
	ID          string `orm:"column(id); size(64); pk"`
	DeviceID    string `orm:"column(deviceid); null; type(text)"`
	Name        string `orm:"column(name); null; type(text)"`
	Description string `orm:"column(description); null; type(text)"`
	State       string `orm:"column(state); null; type(text)"`
	LastOnline  string `orm:"column(last_online); null; type(text)"`
	//TODO: a list of Metadatas

	//Value    string `orm:"column(value);null;type(text)"`
	//Optional bool   `orm:"column(optional);null;type(integ er)"`
	//AttrType string `orm:"column(attr_type);null;type(text)"`
	//Metadata string `orm:"column(metadata);null;type(text)"`
}

//SaveDevice save device
func SaveDevice(doc *Device) error {
	logrus.WithFields(logrus.Fields{
		"Device": doc,
	}).Errorf("b4 Insert err")
	num, err := dbm.DBAccess.Insert(doc)
	logrus.WithFields(logrus.Fields{
		"err": err,
	}).Errorf("Insert err")
	klog.V(4).Infof("Insert affected Num: %d, %v", num, err)
	return err
}

//DeleteDeviceByID delete device by id
func DeleteDeviceByID(id string) error {
	num, err := dbm.DBAccess.QueryTable(DeviceTableName).Filter("id", id).Delete()
	if err != nil {
		klog.Errorf("Something wrong when deleting data: %v", err)
		return err
	}
	klog.V(4).Infof("Delete affected Num: %d", num)
	return nil
}

// UpdateDeviceField update special field
func UpdateDeviceField(deviceID string, col string, value interface{}) error {
	num, err := dbm.DBAccess.QueryTable(DeviceTableName).Filter("id", deviceID).Update(map[string]interface{}{col: value})
	klog.V(4).Infof("Update affected Num: %d, %s", num, err)
	return err
}

// UpdateDeviceFields update special fields
func UpdateDeviceFields(deviceID string, cols map[string]interface{}) error {
	num, err := dbm.DBAccess.QueryTable(DeviceTableName).Filter("id", deviceID).Update(cols)
	klog.V(4).Infof("Update affected Num: %d, %s", num, err)
	return err
}

// QueryDevice query Device
func QueryDevice(key string, condition string) (*[]Device, error) {
	devices := new([]Device)
	_, err := dbm.DBAccess.QueryTable(DeviceTableName).Filter(key, condition).All(devices)
	if err != nil {
		return nil, err
	}
	return devices, nil
}

// QueryDeviceAll query twin
func QueryDeviceAll() (*[]Device, error) {
	devices := new([]Device)
	_, err := dbm.DBAccess.QueryTable(DeviceTableName).All(devices)
	if err != nil {
		return nil, err
	}
	return devices, nil
}

//DeviceUpdate the struct for updating device
type DeviceUpdate struct {
	DeviceID string
	Cols     map[string]interface{}
}

//UpdateDeviceMulti update device  multi
func UpdateDeviceMulti(updates []DeviceUpdate) error {
	var err error
	for _, update := range updates {
		err = UpdateDeviceFields(update.DeviceID, update.Cols)
		if err != nil {
			return err
		}
	}
	return nil
}

//AddDeviceTrans the transaction of add device
func AddDeviceTrans(adds []Device) error {
	logrus.WithFields(logrus.Fields{
		"err": "err",
	}).Errorf("transaction operate func")
	var err error
	obm := dbm.DBAccess
	errx := obm.Begin()

	logrus.WithFields(logrus.Fields{
		"err":  errx,
		"adds": adds,
	}).Errorf("obm begin err")

	for _, add := range adds {
		err = SaveDevice(&add)
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Errorf("SaveDevice err")
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err": err,
			}).Errorf("transaction operate err and rollback")
			klog.Errorf("save device failed: %v", err)
			obm.Rollback()
			return err
		}
	}
	errcom := obm.Commit()
	logrus.WithFields(logrus.Fields{
		"err": errcom,
	}).Errorf("obm commit err")
	return nil
}

//DeleteDeviceTrans the transaction of delete device
func DeleteDeviceTrans(deletes []string) error {
	var err error
	obm := dbm.DBAccess
	obm.Begin()
	for _, delete := range deletes {
		err = DeleteDeviceByID(delete)
		if err != nil {
			obm.Rollback()
			return err
		}
	}
	obm.Commit()
	return nil
}
