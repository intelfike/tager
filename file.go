package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/intelfike/nestmap"
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
	Use:   "file",
	Short: "ファイルを一覧する",
	Long:  "ファイルを一覧する",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.ParseFlags(args)
		if len(args) == 0 {
			addUsage(cmd, " TAG")
			cmd.Help()
			return
		}

		files := make([][]string, 0)
		for _, v := range args {
			cur := rootTags.Child(v)
			if !cur.Exists() {
				fmt.Println(v, "そのようなタグは存在しません")
				continue
			}
			files = append(files, make([]string, 0))
			if cur.IsMap() {
				if cur.HasChild("files") {
					files[len(files)-1] = cur.Child("files").Keys()
				}
				if *showFlagR {
					recNestTag(cur, "", func(nm *nestmap.Nestmap, path string) {
						if !nm.HasChild("files") {
							return
						}
						files[len(files)-1] = nm.Child("files").Keys()
					})
				}
			}
		}
		if len(files) == 0 {
			return
		}
		and := make([]string, len(files[0]))
		copy(and, files[0])
		if len(files) != 1 {
			for _, v := range files[1:] {
				and = andStrings(and, v)
			}
		}
		fmt.Println(strings.Join(and, "\n"))
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

var addFilesCmd = &cobra.Command{
	Use:   "file",
	Short: "タグにファイルを登録する",
	Long:  "タグにファイルを登録する\n登録先のタグが create されている必要があります",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) <= 1 {
			addUsage(cmd, " TAG FILES...")
			cmd.Help()
			return
		}
		cur := rootTags.Child(args[0])
		if !cur.Exists() {
			fmt.Println(args[0], "そのようなタグは存在しません")
			return
		}
		for _, v := range args[1:] {
			_, err := os.Stat(v)
			if err != nil {
				fmt.Println(v, "そのようなファイルは存在しません")
				continue
			}
			full, err := filepath.Abs(v)
			if err != nil {
				fmt.Println(v, "ファイル名の指定が正しくありません")
				continue
			}
			if cur.Child("files").HasChild(full) {
				fmt.Println(v, "というファイルは既に", args[0], "に登録されています")
				continue
			}
			cur.Child("files", full).Set(v)
		}
		if err := save(); err != nil {
			fmt.Println(err)
			return
		}

	},
}

// 削除済みのファイルを削除できない！
var removeFilesCmd = &cobra.Command{
	Use:   "file",
	Short: "タグからファイルの登録を削除する",
	Long:  "タグからファイルの登録を削除する\n削除するファイル名が存在していない場合は無視されます",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) <= 1 {
			addUsage(cmd, " TAG FILES...")
			cmd.Help()
			return
		}
		cur := rootTags.Child(args[0], "files")
		if !cur.Exists() {
			fmt.Println(args[0], "そのようなタグは存在しません")
			return
		}
		for _, v := range args[1:] {
			_, err := os.Stat(v)
			if err != nil {
				fmt.Println(v, "そのようなファイルは存在しません")
				continue
			}
			full, err := filepath.Abs(v)
			if err != nil {
				fmt.Println(v, "ファイル名の指定が正しくありません")
				continue
			}
			cur.Child(full).Remove()
		}
		if err := save(); err != nil {
			fmt.Println(err)
			return
		}

	},
}

var autoremoveFilesCmd = &cobra.Command{
	Use:   "file",
	Short: "タグから存在しないファイルを自動削除する",
	Long:  "タグから存在しないファイルを自動削除する\nタグ名が未指定の場合はすべてのタグが対象です",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// 引数がなければすべてが対象
			args = rootTags.Keys()
		}
		for _, v := range args {
			files := rootTags.Child(v, "files")
			if !files.Exists() {
				fmt.Println(v, "そのようなタグはありません")
				continue
			}

			for _, file := range files.Keys() {
				if _, err := os.Stat(file); err == nil {
					continue
				}
				files.Child(file).Remove()
				fmt.Println(v, "から", file, "というファイルを削除しました")
			}
		}
		if err := save(); err != nil {
			fmt.Println(err)
			return
		}
	},
}
