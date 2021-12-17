package cdmanager

//func (cdm *CDManager) runCDManager() {
//	go func() {
//		for {
//			select {
//			case <-beehiveContext.Done():
//				klog.Warning("Stop DeviceManager Receiving.")
//				logrus.WithFields(logrus.Fields{
//					"module": "dmanager",
//					"func":   "runDeviceManager()",
//				}).Infof("Stop DeviceManager Receiving.")
//
//				return
//			default:
//			}
//			if msg, ok := beehiveContext.Receive("cdmgr"); ok == nil {
//				klog.Info("DeviceManager receive msg")
//				logrus.WithFields(logrus.Fields{
//					"msg":    msg,
//					"module": "CDManager",
//					"func":   "runcCDeviceManager()",
//				}).Infof("CDeviceManager receive msg.")
//
//				err := cdm.distributeMsg(msg)
//				if err != nil {
//					klog.Info("distributeMsg failed.")
//					logrus.WithFields(logrus.Fields{
//						"module": "CDManager",
//						"func":   "runCDeviceManager()",
//					}).Infof("distributeMsg failed.")
//
//				}
//			}
//		}
//	}()
//}
//
//func (cdm *CDManager) distributeMsg(m interface{}) error {
//	klog.Info("cDmanager Received message: %v", m)
//	msg, ok := m.(model.Message)
//	if !ok {
//		logrus.WithFields(logrus.Fields{
//			"message": msg,
//		}).Infof("Cloud DManager Got Message.")
//	}
//	return nil
//}
