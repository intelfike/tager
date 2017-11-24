package main

import (
	"fmt"
	"os"

	"github.com/intelfike/nestmap"
	"github.com/spf13/cobra"
)

var showTagsCmd = &cobra.Command{
	Use:   "tag",
	Short: "タグを一覧する",
	Long:  "タグを一覧する",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.ParseFlags(args)
		if len(args) == 0 {
			if *showFlagR {
				fmt.Println("-r --recursive を指定した場合には表示するタグ名も入力してください")
				return
			}
			showTags(rootTags.Keys())
			return
		}

		cur := rootTags.Child(args[0])
		if !cur.Exists() {
			fmt.Println("そのようなタグは存在しません")
			return
		}
		if cur.Child("tags").IsMap() {
			if *showFlagR {
				recNestTag(cur, args[0], func(nm *nestmap.Nestmap, path string) {
					fmt.Println(path + "/" + nm.BottomPath().(string))
				})
			} else {
				showTags(cur.Child("tags").Keys())
			}
		}
	},
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "新しいタグを作成する",
	Long:  "新しいタグを作成する",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			addUsage(cmd, " TAG")
			cmd.Help()
			return
		}
		root := rootTags
		for _, v := range args {
			if root.HasChild(v) {
				fmt.Println(v, "というタグは既に存在しています")
				continue
			}
			// タグの初期化
			root.Child(v).MakeMap()
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
			addUsage(cmd, " TAG")
			cmd.Help()
			return
		}
		root := rootTags
		for _, v := range args {
			if !root.HasChild(v) {
				fmt.Println(v, "というタグは存在しません")
				continue
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

// 循環参照のチェックをだね
var addTagsCmd = &cobra.Command{
	Use:   "tag",
	Short: "タグにタグを登録する",
	Long:  "タグにタグを登録する\n登録先のタグ、登録するタグの両方が create されている必要があります",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) <= 1 {
			addUsage(cmd, " TAG TAGS...")
			cmd.Help()
			return
		}
		cur := rootTags.Child(args[0])
		if !cur.Exists() {
			fmt.Println(args[0], "そのようなタグは存在しません")
			return
		}
		for _, v := range args[1:] {
			if !rootTags.HasChild(v) {
				fmt.Println(v, "そのようなタグは存在しません")
				continue
			}
			if args[0] == v {
				fmt.Println(v, "登録元と登録先のタグが同じです")
				continue
			}
			if cur.Child("tags").HasChild(v) {
				fmt.Println(v, "というタグは既に", args[0], "に登録されています")
				continue
			}
			cur.Child("tags", v).Set(v)
		}
		// 循環参照をチェックして拒否するため
		recNestTag(cur, "", func(nm *nestmap.Nestmap, path string) {
			s := nm.BottomPath().(string)
			unique := args[0] != s
			if !unique {
				fmt.Println(s, ":循環参照です\n登録に失敗しました")
				os.Exit(1)
			}
		})
		if err := save(); err != nil {
			fmt.Println(err)
			return
		}
	},
}

// ==================== remove ====================

var removeTagsCmd = &cobra.Command{
	Use:   "tag",
	Short: "タグからタグを削除する",
	Long:  "タグからタグを削除する\n削除するタグが存在していない場合は無視されます",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) <= 1 {
			addUsage(cmd, " TAG TAGS...")
			cmd.Help()
			return
		}
		cur := rootTags.Child(args[0], "tags")
		if !cur.Exists() {
			fmt.Println(args[0], "そのようなタグは存在しません")
			return
		}
		for _, v := range args[1:] {
			if !rootTags.HasChild(v) {
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

// ==================== autoremove ====================
var autoremoveTagsCmd = &cobra.Command{
	Use:   "tag",
	Short: "タグから存在しないタグを自動削除する",
	Long:  "タグから存在しないタグを自動削除する\nタグ名が未指定の場合はすべてのタグが対象です",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// 引数がなければすべてが対象
			args = rootTags.Keys()
		}
		for _, v := range args {
			tags := rootTags.Child(v, "tags")
			if !tags.Exists() {
				fmt.Println(v, "そのようなタグはありません")
				continue
			}
			for _, tag := range tags.Keys() {
				if rootTags.HasChild(tag) {
					continue
				}
				tags.Child(tag).Remove()
				fmt.Println(v, "から", tag, "というタグを削除しました")
			}
		}
		if err := save(); err != nil {
			fmt.Println(err)
			return
		}

	},
}
