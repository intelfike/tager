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
	Use:   "ls",
	Short: "ファイルを一覧する",
	Long:  "ファイルを一覧する",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.SetUsageTemplate("tager file ls [tags]...")
			cmd.Help()
			return
		}
		cur := config.Child("root", "tags", args[0], "files")
		if !cur.Exists() {
			fmt.Println("そのようなタグは存在しません")
			return
		}
		if cur.IsMap() {
			fmt.Println(strings.Join(cur.Keys(), "\n"))
		}
	},
}

// ==================== add ====================

var fileAddCmd = &cobra.Command{
	Use:   "add",
	Short: "ファイルにタグを付与する",
	Long:  "ファイルにタグを付与する\nタグが存在しない場合は登録できません",
	Run: func(cmd *cobra.Command, args []string) {

	},
}
