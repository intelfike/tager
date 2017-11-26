package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var showAllCmd = &cobra.Command{
	Use:   "all [flags] TAG",
	Short: "タグ内のすべてのデータを表示する",
	Long:  "タグ内のすべてのデータを表示する",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}
		showCommentCmd.Run(cmd, args)
		fmt.Println()
		fmt.Println("tags:")
		showTagsCmd.Run(cmd, args)
		fmt.Println()
		fmt.Println("files:")
		showFilesCmd.Run(cmd, args)
	},
}

var autoremoveAllCmd = &cobra.Command{
	Use:   "all [TAG]",
	Short: "タグから存在しないタグとファイルを自動削除する",
	Long:  "タグから存在しないタグとファイルを自動削除する\nタグ名が未指定の場合はすべてのタグが対象です",
	Run: func(cmd *cobra.Command, args []string) {
		autoremoveTagsCmd.Run(cmd, args)
		autoremoveFilesCmd.Run(cmd, args)
	},
}
