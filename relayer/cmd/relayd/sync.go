package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/stratosnet/sds/framework/utils"

	"github.com/stratosnet/sds/relayer/cmd/relayd/setting"
	"github.com/stratosnet/sds/relayer/rpc"
	"github.com/stratosnet/sds/relayer/server"
)

func sync(cmd *cobra.Command, args []string) error {
	if len(args) != 1 || len(args[0]) == 0 {
		utils.ErrorLog("wrong number of arguments")
		return nil
	}
	txHash := args[0]
	// verify if wallet and public key match
	if len(txHash) == 0 {
		utils.Logf("empty txhash")
		return errors.New("empty txhash")
	}

	c, err := rpc.Dial(setting.IpcEndpoint)
	if err != nil {
		utils.ErrorLog(err)
		return err
	}
	defer c.Close()

	callRpc(c, "sync", args)
	return nil
}

func callRpc(c *rpc.Client, line string, param []string) bool {
	var result server.CmdResult

	err := c.Call(&result, "relayer_"+line, param)
	if err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Println(result.Msg)
	return true
}

func syncPreRunE(cmd *cobra.Command, _ []string) error {
	homePath, err := cmd.Flags().GetString(server.Home)
	if err != nil {
		utils.ErrorLog("failed to get 'home' path for the relayd process")
		return err
	}
	homePath, err = utils.Absolute(homePath)
	if err != nil {
		utils.ErrorLog("cannot convert home path to absolute path")
		return err
	}
	setting.HomePath = homePath
	setting.SetIPCEndpoint(homePath)
	return nil
}
