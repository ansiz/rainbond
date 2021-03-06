// RAINBOND, Application Management Platform
// Copyright (C) 2014-2017 Goodrain Co., Ltd.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. For any non-GPL usage of Rainbond,
// one or multiple Commercial Licenses authorized by Goodrain Co., Ltd.
// must be obtained first.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package controller

import (
	"github.com/Sirupsen/logrus"
	v1 "github.com/goodrain/rainbond/worker/appm/types/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type applyConfigController struct {
	controllerID string
	appService   v1.AppService
	manager      *Manager
	stopChan     chan struct{}
}

// Begin begins applying rule
func (a *applyConfigController) Begin() {
	nowApp := a.manager.store.GetAppService(a.appService.ServiceID)
	nowConfigMaps := nowApp.GetConfigMaps()
	newConfigMaps := a.appService.GetConfigMaps()
	var nowConfigMapUpdate = make(map[string]bool, len(nowConfigMaps))
	for _, now := range nowConfigMaps {
		nowConfigMapUpdate[now.Name] = false
	}
	for _, new := range newConfigMaps {
		newc, err := a.manager.client.CoreV1().ConfigMaps(nowApp.TenantID).Update(new)
		if err != nil {
			logrus.Errorf("update config map failure %s", err.Error())
		}
		nowApp.SetConfigMap(newc)
		if _, ok := nowConfigMapUpdate[new.Name]; ok {
			nowConfigMapUpdate[new.Name] = true
		}
	}
	for name, handle := range nowConfigMapUpdate {
		if !handle {
			if err := a.manager.client.CoreV1().ConfigMaps(nowApp.TenantID).Delete(name, &metav1.DeleteOptions{}); err != nil {
				logrus.Errorf("delete config map failure %s", err.Error())
			}
		}
	}
	a.manager.callback(a.controllerID, nil)
}

func (a *applyConfigController) Stop() error {
	close(a.stopChan)
	return nil
}
