package wechat

import (
	"errors"
	"sync"
)

type (
	//应用管理器
	WechatAppManager struct {
		Apps map[string]*WechatApp
		sync.RWMutex
		CallBackUri string
	}
)

//实例化应用管理器
//callBackUri表示回调地址的第一维uri。以：http://www.xx.com/callback/code_121j12jsusds 为例，callback是第一维，code_121j12jsusds是
//第二维。实际上，我们的后台回调地址，必须设置为此种形式，以便支持一个程序处理对个微信公众号的请求
func NewWechatAppManager(callBackUri ...string) *WechatAppManager {
	uri := ""
	if len(callBackUri) == 0 {
		uri = "/callback/"
	} else {
		uri = callBackUri[0]
	}

	return &WechatAppManager{Apps: make(map[string]*WechatApp), CallBackUri: uri}
}

//添加一个应用到应用管理器
func (m *WechatAppManager) AddApp(app *WechatApp) (bool, error) {
	var done bool = false
	var err error = nil
	//check params
	if app == nil || app.Configure.AppId == "" {
		err = errors.New("无效的应用编号")
		return done, err
	}
	//read lock
	m.RLock()
	_, exists := m.Apps[app.Configure.AppId]
	if exists {
		m.RUnlock()
		err = errors.New("同样的应用id下，已存在一个应用，无法添加，appId :" + app.Configure.AppId)
	} else {
		m.RUnlock()
		m.Lock()
		if _, exists = m.Apps[app.Configure.AppId]; !exists {
			m.Apps[app.Configure.AppId] = app
		} else {
			err = errors.New("同样的应用id下，已存在一个应用，无法添加，appId :" + app.Configure.AppId)
		}
		m.Unlock()
	}
	return done, err
}

//从应用管理器中，移除一个应用
func (m *WechatAppManager) RemoveApp(appId string) (bool, error) {
	var done bool = false
	var err error = nil
	//参数检测
	if appId == "" {
		err = errors.New("无效的应用id")
		return done, err
	}
	//read lock
	m.RLock()
	_, exists := m.Apps[appId]
	if !exists {
		m.RUnlock()
		err = errors.New("同样的应用id下，已存在一个应用，无法添加，appId :" + appId)
	} else {
		m.RUnlock()
		m.Lock()
		if _, exists = m.Apps[appId]; exists {
			delete(m.Apps, appId)
			done = true
		} else {
			err = errors.New("同样的应用id下，已存在一个应用，无法添加，appId :" + appId)
		}
		m.Unlock()
	}
	return done, err
}

//更新一个应用
func (m *WechatAppManager) UpdateApp(appId string, app *WechatApp) (bool, error) {
	var done bool = false
	var err error = nil
	//check params
	if app == nil || appId == "" {
		err = errors.New("更新失败，无效的应用或id")
		return done, err
	}
	//read lock
	m.RLock()
	_, exists := m.Apps[appId]
	if !exists {
		m.RUnlock()
		err = errors.New("同样的应用id下，已存在一个应用，无法添加，appId :" + appId)
	} else {
		m.RUnlock()
		m.Lock()
		if _, exists := m.Apps[appId]; exists {
			m.Apps[appId] = app
		} else {
			err = errors.New("同样的应用id下，已存在一个应用，无法添加，appId :" + appId)
		}
		m.Unlock()
	}
	return done, err
}

//从管理器中，获取一个应用
func (m *WechatAppManager) GetApp(appId string) (*WechatApp, bool) {
	m.RLock()
	app, exists := m.Apps[appId]
	m.RUnlock()
	return app, exists
}
