package p2pserver

import (
	"context"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/stratosnet/sds/framework/client/cf"
	"github.com/stratosnet/sds/framework/core"
	"github.com/stratosnet/sds/framework/msg"
	"github.com/stratosnet/sds/framework/utils"
	"github.com/stratosnet/sds/pp/setting"
)

type offline struct {
	IsSp           bool
	NetworkAddress string
}

type lastReconnectRecord struct {
	SpP2PAddress                string
	Time                        time.Time
	NextAllowableReconnectInSec int64
}

func (p *P2pServer) initClient() {
	p.offlineChan = make(chan *offline, 2)
	p.cachedConnMap = &sync.Map{}
	p.connMap = make(map[string]*cf.ClientConn)
}

func (p *P2pServer) SetOptionFunctions(write func(context.Context, *msg.RelayMsgBuf), read func(*msg.RelayMsgBuf),
	handle func(context.Context, *msg.RelayMsgBuf)) {
	p.onWriteFunc = write
	p.onReadFunc = read
	p.onHandleFunc = handle
}

func (p *P2pServer) NewClientToMainSp(ctx context.Context, server string) error {
	utils.DebugLog("NewClientToMainSp: to", server, " hb: true, rec: true")
	_, err := p.newClient(ctx, server, true, false, true)
	return err
}

func (p *P2pServer) NewClientToAlternativeSp(ctx context.Context, server string) (*cf.ClientConn, error) {
	utils.DebugLog("NewClientToAlternativeSp: to", server)
	return p.newClient(ctx, server, false, false, false)
}

func (p *P2pServer) NewClientToPp(ctx context.Context, server string, heartbeat bool) (*cf.ClientConn, error) {
	utils.DebugLog("NewClientToPp: to", server)
	return p.newClient(ctx, server, heartbeat, false, false)
}

func (p *P2pServer) newClient(ctx context.Context, server string, heartbeat, reconnect, spconn bool) (*cf.ClientConn, error) {
	onConnect := cf.OnConnectOption(func(c core.WriteCloser) bool {
		utils.DebugLog("on connect")
		return true
	})
	onError := cf.OnErrorOption(func(c core.WriteCloser) {
		utils.Log("on error")
	})
	onClose := cf.OnCloseOption(func(c core.WriteCloser) {
		cc, ok := c.(*cf.ClientConn)
		if !ok {
			utils.DebugLogf("Closing client conn %v. The conn object is invalid", server)
			return
		}
		if cc == nil {
			utils.DebugLogf("Closing client conn %v. The conn object is nil", server)
			return
		}

		utils.Log("on close", cc.GetName())
		p.clientMutex.Lock()
		delete(p.connMap, cc.GetName())
		p.clientMutex.Unlock()

		offlineInfo := &offline{
			IsSp:           false,
			NetworkAddress: cc.GetRemoteAddr(),
		}
		if p.SpConnValid() && p.mainSpConn.GetName() == cc.GetName() {
			utils.DebugLog("lost SP conn, name: ", p.mainSpConn.GetName(), " netId is ", p.mainSpConn.GetNetID())
			p.mainSpConn = nil
			offlineInfo.IsSp = true
		}

		select {
		case p.offlineChan <- offlineInfo:
		default:
			break
		}
	})

	serverPort, err := strconv.ParseUint(setting.Config.Node.Connectivity.NetworkPort, 10, 16)
	if err != nil {
		return nil, errors.Wrapf(err, "Invalid port number in config [%v]", setting.Config.Node.Connectivity.NetworkPort)
	}
	serverPortOpt := cf.ServerPortOption(uint16(serverPort))

	var ckv []cf.ContextKV
	for _, key := range p.connContextKey {
		ckv = append(ckv, cf.ContextKV{Key: key, Value: ctx.Value(key)})
	}

	options := []cf.ClientOption{
		onConnect,
		onError,
		onClose,
		cf.OnWriteOption(p.onWriteFunc),
		cf.OnHandleOption(p.onHandleFunc),
		cf.BufferSizeOption(100),
		cf.ReconnectOption(reconnect),
		cf.HeartCloseOption(!heartbeat),
		cf.LogOpenOption(true),
		cf.MinAppVersionOption(setting.Config.Version.MinAppVer),
		cf.P2pAddressOption(p.GetP2PAddress().String()),
		cf.ServerIpOption(setting.NetworkIP),
		serverPortOpt,
		cf.ContextKVOption(ckv),
	}
	conn := cf.CreateClientConn(server, options...)

	// setting p.mainSpConn earlier than calling conn.Start() to avoid race condition
	if spconn {
		p.mainSpConn = conn
	}
	p.clientMutex.Lock()
	p.connMap[server] = conn
	p.clientMutex.Unlock()

	conn.Start()
	return conn, nil
}

func (p *P2pServer) GetConnectionName(conn core.WriteCloser) string {
	if conn == nil {
		return ""
	}
	switch conn := conn.(type) {
	case *core.ServerConn:
		return conn.GetName()
	case *cf.ClientConn:
		return conn.GetName()
	}
	return ""
}

func (p *P2pServer) GetClientConn(networkAddr string) (*cf.ClientConn, bool) {
	p.clientMutex.Lock()
	defer p.clientMutex.Unlock()
	if cc, ok := p.connMap[networkAddr]; ok {
		return cc, true
	} else {
		return nil, false
	}
}

func (p *P2pServer) CleanUpConnMap(fileHash string) {
	p.cachedConnMap.Range(func(k, v interface{}) bool {
		if strings.HasPrefix(k.(string), fileHash) {
			p.DeleteConnFromCache(k.(string))
		}
		return true
	})
}

func (p *P2pServer) ReadOfflineChan() chan *offline {
	return p.offlineChan
}

func (p *P2pServer) SpConnValid() bool {
	return p.mainSpConn != nil
}

func (p *P2pServer) GetSpName() string {
	if !p.SpConnValid() {
		return "{NA} Invalid SpConn"
	}
	return p.mainSpConn.GetName()
}

// StoreConnToCache access function for member cachedConnMap
func (p *P2pServer) StoreConnToCache(key string, conn *cf.ClientConn) {
	p.cachedConnMap.Store(key, conn)
}

// LoadConnFromCache access function for member cachedConnMap
func (p *P2pServer) LoadConnFromCache(key string) (*cf.ClientConn, bool) {
	if c, ok := p.cachedConnMap.Load(key); ok {
		return c.(*cf.ClientConn), true
	} else {
		return nil, false
	}
}

// DeleteConnFromCache access function for member cachedConnMap
func (p *P2pServer) DeleteConnFromCache(key string) {
	p.cachedConnMap.Delete(key)
}

func (p *P2pServer) RangeCachedConn(prefix string, rf func(k, v interface{}) bool) {
	p.cachedConnMap.Range(
		func(k, v interface{}) bool {
			if strings.HasPrefix(k.(string), prefix) {
				return rf(k, v)
			}
			return true
		})
	p.cachedConnMap.Range(rf)
}

func (p *P2pServer) GetSpConn() *cf.ClientConn {
	return p.mainSpConn
}

// RecordSpMaintenance return boolean flag of switching to new SP
func (p *P2pServer) RecordSpMaintenance(spP2pAddress string, recordTime time.Time) bool {
	if value, ok := p.SPMaintenanceMap.Load(spP2pAddress); ok {
		lastRecord := value.(lastReconnectRecord)
		if !lastRecord.canReconnect() {
			return false
		}

		nextReconnectInterval := int64(math.Min(MAX_RECONNECT_INTERVAL_THRESHOLD,
			float64(lastRecord.NextAllowableReconnectInSec*RECONNECT_INTERVAL_MULTIPLIER)))
		p.SPMaintenanceMap.Store(spP2pAddress, lastReconnectRecord{
			SpP2PAddress:                spP2pAddress,
			Time:                        recordTime,
			NextAllowableReconnectInSec: nextReconnectInterval,
		})
		return true
	}

	p.SPMaintenanceMap.Store(spP2pAddress, lastReconnectRecord{
		SpP2PAddress:                spP2pAddress,
		Time:                        recordTime,
		NextAllowableReconnectInSec: MIN_RECONNECT_INTERVAL_THRESHOLD,
	})
	return true
}

func (p *P2pServer) CanConnectToSp(p2pAddress string) bool {
	if value, ok := p.SPMaintenanceMap.Load(p2pAddress); ok {
		lastRecord := value.(lastReconnectRecord)
		if lastRecord.SpP2PAddress == p2pAddress && !lastRecord.canReconnect() {
			return false
		}
	}

	return true
}

func (l lastReconnectRecord) canReconnect() bool {
	return time.Now().After(l.Time.Add(time.Duration(l.NextAllowableReconnectInSec) * time.Second))
}
