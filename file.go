package main

import (
	"fmt"
	"path/filepath"
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

var showFilesCmd = &cobra.Command{
	Use:   "file [flags] TAG...",
	Short: "ファイルを一覧する",
	Long:  "ファイルを一覧する\n複数のタグを指定した場合はAND計算をします",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}
		ss, err := tager.getFilesAND(args...)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(strings.Join(ss, "\n"))
	},
}

// ==================== add ====================

var addFilesCmd = &cobra.Command{
	Use:   "file [flags] TAG FILES...",
	Short: "タグにファイルを登録する",
	Long:  "タグにファイルを登録する\n登録先のタグが create されている必要があります",
	Run: func(cmd *cobra.Command, args []string) {
		cur := rootTags.Child(args[0])
		if !cur.Exists() {
			fmt.Println(args[0], "そのようなタグは存在しません")
			return
		}
		fmt.Println(args)
		if *addFileFlagR {
			tager.tagAddFileRec(args[0], args[1:]...)
		} else {
			tager.tagAddFile(args[0], args[1:]...)
		}
		tager.saveConfig()
	},
}

// 削除済みのファイルを削除できない！
var removeFilesCmd = &cobra.Command{
	Use:   "file [flags] TAG FILES...",
	Short: "タグからファイルの登録を削除する",
	Long:  "タグからファイルの登録を削除する\n削除するファイル名が存在していない場合は無視されます",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) <= 1 {
			cmd.Help()
			return
		}
		cur := rootTags.Child(args[0])
		if !cur.Exists() {
			fmt.Println(args[0], "そのようなタグは存在しません")
			return
		}
		for _, v := range args[1:] {
			if !fileExists(v) {
				fmt.Println(v, "そのようなファイルは存在しません")
				continue
			}
			full, err := filepath.Abs(v)
			if err != nil {
				fmt.Println(v, "ファイル名の指定が正しくありません")
				continue
			}
			cur.Child("files", full).Remove()
		}
	},
}

var autoremoveFilesCmd = &cobra.Command{
	Use:   "file [TAG]...",
	Short: "タグから存在しないファイルを自動削除する",
	Long:  "タグから存在しないファイルを自動削除する\nタグ名が未指定の場合はすべてのタグが対象です",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// 引数がなければすべてが対象
			args = rootTags.Keys()
		}
		for _, v := range args {
			cur := rootTags.Child(v)
			if !cur.Exists() {
				fmt.Println(v, "そのようなタグはありません")
				continue
			}
			if !cur.HasChild("files") {
				continue
			}
			for _, file := range cur.Child("files").Keys() {
				if !fileExists(file) {
					continue
				}
				cur.Child("files", file).Remove()
				fmt.Println(v, "から", file, "というファイルを削除しました")
			}
		}
	},
}
