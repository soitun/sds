package event

import (
	"context"

	"github.com/stratosnet/sds/framework/core"
	"github.com/stratosnet/sds/msg/header"
	"github.com/stratosnet/sds/msg/protos"
	"github.com/stratosnet/sds/pp"
	"github.com/stratosnet/sds/pp/peers"
	"github.com/stratosnet/sds/pp/requests"
)

// StartMaintenance sends a request to SP to temporarily put the current node into maintenance mode
func StartMaintenance(ctx context.Context, duration uint64) error {
	req := requests.ReqStartMaintenance(duration)
	pp.Log(ctx, "Sending maintenance start request to SP!")
	peers.SendMessageToSPServer(ctx, req, header.ReqStartMaintenance)
	return nil
}

func StopMaintenance(ctx context.Context) error {
	req := requests.ReqStopMaintenance()
	pp.Log(ctx, "Sending maintenance stop request to SP!")
	peers.SendMessageToSPServer(ctx, req, header.ReqStopMaintenance)
	return nil
}

func RspStartMaintenance(ctx context.Context, _ core.WriteCloser) {
	var target protos.RspStartMaintenance
	if !requests.UnmarshalData(ctx, &target) {
		pp.DebugLog(ctx, "Cannot unmarshal start maintenance response")
		return
	}

	if target.Result.State != protos.ResultState_RES_SUCCESS {
		pp.Logf(ctx, "cannot start maintenance: %v", target.Result.Msg)
		return
	}
}

func RspStopMaintenance(ctx context.Context, _ core.WriteCloser) {
	var target protos.RspStopMaintenance
	if !requests.UnmarshalData(ctx, &target) {
		pp.DebugLog(ctx, "Cannot unmarshal stop maintenance response")
		return
	}

	if target.Result.State != protos.ResultState_RES_SUCCESS {
		pp.Logf(ctx, "failed to stop maintenance: %v", target.Result.Msg)
		return
	}
}
