package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "タグ関連のコマンド",
	Long:  "タグを管理するためのサブコマンド",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var taglsCmd = &cobra.Command{
	Use:   "ls",
	Short: "タグを一覧する",
	Long:  "タグを一覧する",
	Run: func(cmd *cobra.Command, args []string) {
		cur := moveTag(args).Child("tags")
		if !cur.Exists() {
			fmt.Println("そのようなタグは存在しません")
			return
		}
		if cur.IsMap() {
			fmt.Println(strings.Join(cur.Keys(), "\n"))
		}
	},
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "新しいタグを作成する",
	Long:  "新しいタグを作成する",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}
		root := config.Child("root", "tags")
		for _, v := range args {
			if root.HasChild(v) {
				fmt.Println(v, "というタグは既に存在しています")
			}
			// タグの初期化
			tagini := getInitedTag()
			root.Child(v).Set(tagini.Interface())
		}
		if err := save(); err != nil {
			fmt.Println(err)
			return
		}
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "タグを完全に削除する",
	Long:  "タグを完全に削除する",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}
		root := config.Child("root", "tags")
		for _, v := range args {
			if !root.HasChild(v) {
				fmt.Println(v, "というタグは存在しません")
			}
			// タグの初期化
			root.Child(v).Remove()
		}
		if err := save(); err != nil {
			fmt.Println(err)
			return
		}
	},
}

// ==================== add ====================

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "タグにデータを登録する",
	Long:  "タグにデータを登録する",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// addするときにリンクだけ貼る？
var addTagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "タグにタグを登録する",
	Long:  "タグにタグを登録する\n登録先のタグ、登録するタグの両方が create されている必要があります",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) <= 1 {
			cmd.SetUsageTemplate("tager tag add tags [tag] [tags]...")
			cmd.Help()
			return
		}
		cur := moveTag(args[0:1]).Child("tags")
		if !cur.Exists() {
			fmt.Println(args[0], "そのようなタグは存在しません")
			return
		}
		for _, v := range args[1:] {
			if !config.Child("root", "tags").HasChild(v) {
				fmt.Println(v, "そのようなタグは存在しません")
				continue
			}
			if args[0] == v {
				fmt.Println(v, "登録元と登録先のタグが同じです")
				return
			}
			cur.Child(v).Set(v)
		}
		if err := save(); err != nil {
			fmt.Println(err)
			return
		}
	},
}

var addFilesCmd = &cobra.Command{
	Use:   "files",
	Short: "タグにファイルを登録する",
	Long:  "タグにファイルを登録する\n登録先のタグが create されている必要があります",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) <= 1 {
			cmd.SetUsageTemplate("tager tag add files [tag] [file]...")
			cmd.Help()
			return
		}
		cur := moveTag(args[0:1]).Child("files")
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
			cur.Child(full).Set(v)
		}
		if err := save(); err != nil {
			fmt.Println(err)
			return
		}

	},
}

// ==================== remove ====================

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "タグからデータを削除する",
	Long:  "タグからデータを削除する\n削除するデータが存在していない場合は無視されます",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var removeTagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "タグからタグを削除する",
	Long:  "タグからタグを削除する\n削除するタグが存在していない場合は無視されます",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) <= 1 {
			cmd.SetUsageTemplate("tager tag add tags [tag] [tags]...")
			cmd.Help()
			return
		}
		cur := moveTag(args[0:1]).Child("tags")
		if !cur.Exists() {
			fmt.Println(args[0], "そのようなタグは存在しません")
			return
		}
		for _, v := range args[1:] {
			if !config.Child("root", "tags").HasChild(v) {
				fmt.Println(v, "そのようなタグは存在しません")
				continue
			}
			cur.Child(v).Remove()
		}
		if err := save(); err != nil {
			fmt.Println(err)
			return
		}

	},
}

// 削除済みのファイルを削除できない！
var removeFilesCmd = &cobra.Command{
	Use:   "files",
	Short: "タグからファイルの登録を削除する",
	Long:  "タグからファイルの登録を削除する\n削除するファイル名が存在していない場合は無視されます",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) <= 1 {
			cmd.SetUsageTemplate("tager tag add files [tag] [file]...")
			cmd.Help()
			return
		}
		cur := moveTag(args[0:1]).Child("files")
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

// ==================== autoremove ====================

var autoremoveCmd = &cobra.Command{
	Use:   "autoremove",
	Short: "タグから存在しないデータを自動削除する",
	Long:  "タグから存在しないデータを自動削除する",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
var autoremoveAllCmd = &cobra.Command{
	Use:   "all",
	Short: "タグから存在しないタグとファイルを自動削除する\nタグ名が未指定の場合はすべてのタグが対象です",
	Long:  "タグから存在しないタグとファイルを自動削除する\nタグ名が未指定の場合はすべてのタグが対象です",
	Run: func(cmd *cobra.Command, args []string) {
		autoremoveTagsCmd.Run(cmd, args)
		autoremoveFilesCmd.Run(cmd, args)
	},
}
var autoremoveTagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "タグから存在しないタグを自動削除する",
	Long:  "タグから存在しないタグを自動削除する\nタグ名が未指定の場合はすべてのタグが対象です",
	Run: func(cmd *cobra.Command, args []string) {
		rootTags := config.Child("root", "tags")
		if len(args) == 0 {
			// 引数がなければすべてが対象
			args = rootTags.Keys()
		}
		for _, v := range args {
			tags := rootTags.Child(v, "tags")
			for _, tag := range tags.Keys() {
				if rootTags.HasChild(tag) {
					continue
				}
				tags.Child(tag).Remove()
			}
		}
		if err := save(); err != nil {
			fmt.Println(err)
			return
		}

	},
}
var autoremoveFilesCmd = &cobra.Command{
	Use:   "files",
	Short: "タグから存在しないファイルを自動削除する",
	Long:  "タグから存在しないファイルを自動削除する\nタグ名が未指定の場合はすべてのタグが対象です",
	Run: func(cmd *cobra.Command, args []string) {
		rootTags := config.Child("root", "tags")
		if len(args) == 0 {
			// 引数がなければすべてが対象
			args = rootTags.Keys()
		}
		for _, v := range args {
			tags := rootTags.Child(v, "tags")
			for _, file := range tags.Keys() {
				if _, err := os.Stat(file); err != nil {
					continue
				}
				tags.Child(file).Remove()
			}
		}
		if err := save(); err != nil {
			fmt.Println(err)
			return
		}
	},
}
