package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "ファイル関連のコマンド",
	Long:  "ファイルを管理するためのサブコマンド",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var filelsCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"files"},
	Short:   "ファイルを一覧する",
	Long:    "ファイルを一覧する",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.SetUsageTemplate("tager file ls [tag]...")
			cmd.Help()
			return
		}
		cur := moveTag(args).Child("files")
		if !cur.Exists() {
			fmt.Println("そのようなタグは存在しません")
			return
		}
		if cur.IsMap() {
			fmt.Println(strings.Join(cur.Keys(), "\n"))
		}
	},
}
