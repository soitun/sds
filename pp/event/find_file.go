package event

import (
	"context"

	"github.com/stratosnet/sds/framework/core"
	"github.com/stratosnet/sds/framework/msg/header"
	"github.com/stratosnet/sds/framework/utils"
	"github.com/stratosnet/sds/pp"
	"github.com/stratosnet/sds/pp/api/rpc"
	"github.com/stratosnet/sds/pp/file"
	"github.com/stratosnet/sds/pp/p2pserver"
	"github.com/stratosnet/sds/pp/requests"
	"github.com/stratosnet/sds/pp/setting"
	"github.com/stratosnet/sds/sds-msg/protos"
)

func FindFileList(ctx context.Context, fileName string, walletAddr string, pageId uint64, keyword string, fileType int,
	isUp bool, walletPubkey, wsign []byte, reqTime int64) {
	if setting.CheckLogin() {
		p2pserver.GetP2pServer(ctx).SendMessageToSPServer(
			ctx,
			requests.FindFileListData(
				fileName, walletAddr, p2pserver.GetP2pServer(ctx).GetP2PAddress().String(),
				pageId, keyword, protos.FileSortType(fileType), isUp, walletPubkey, wsign, reqTime,
			),
			header.ReqFindMyFileList,
		)
	}
}

func RspFindMyFileList(ctx context.Context, conn core.WriteCloser) {
	pp.DebugLog(ctx, "get RspFindMyFileList")
	var target protos.RspFindMyFileList
	if err := VerifyMessage(ctx, header.RspFindMyFileList, &target); err != nil {
		utils.ErrorLog("failed verifying the message, ", err.Error())
		return
	}
	rpcResult := &rpc.FileListResult{}

	// fail to unmarshal data, not able to determine if and which RPC client this is from, let the client timeout
	if !requests.UnmarshalData(ctx, &target) {
		return
	}

	// serv the RPC user when the ReqId is not empty
	reqId := core.GetRemoteReqId(ctx)
	if reqId != "" {
		defer file.SetFileListResult(target.WalletAddress+reqId, rpcResult)
	}

	if target.P2PAddress != p2pserver.GetP2pServer(ctx).GetP2PAddress().String() {
		rpcResult.Return = rpc.WRONG_PP_ADDRESS
		return
	}

	if target.Result.State != protos.ResultState_RES_SUCCESS {
		utils.ErrorLog(target.Result.Msg)
		rpcResult.Return = rpc.INTERNAL_COMM_FAILURE
		return
	}

	if len(target.FileInfo) == 0 {
		pp.Log(ctx, "There are no files stored")
		rpcResult.Return = rpc.SUCCESS
		rpcResult.TotalNumber = target.TotalFileNumber
		rpcResult.PageId = target.PageId
		return
	}

	var fileInfos = make([]rpc.FileInfo, 0)
	for _, info := range target.FileInfo {
		pp.Log(ctx, "_______________________________")
		if info.IsDirectory {
			pp.Log(ctx, "Directory name:", info.FileName)
			pp.Log(ctx, "Directory hash:", info.FileHash)
		} else {
			pp.Log(ctx, "File name:", info.FileName)
			pp.Log(ctx, "File hash:", info.FileHash)
		}
		pp.Log(ctx, "CreateTime :", info.CreateTime)
		fileInfos = append(fileInfos, rpc.FileInfo{
			FileHash:   info.FileHash,
			FileSize:   info.FileSize,
			FileName:   info.FileName,
			CreateTime: info.CreateTime,
		})
	}

	pp.Log(ctx, "===============================")
	pp.Logf(ctx, "Total: %d  Page: %d", target.TotalFileNumber, target.PageId)

	rpcResult.Return = rpc.SUCCESS
	rpcResult.TotalNumber = target.TotalFileNumber
	rpcResult.PageId = target.PageId
	rpcResult.FileInfo = fileInfos
}
