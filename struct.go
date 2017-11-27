package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/intelfike/nestmap"
)

type Tager struct {
	config   *nestmap.Nestmap
	rootTags *nestmap.Nestmap
}

func (t *Tager) readConfig(filename string) {
	dir, _ := filepath.Split(filename)
	// 設定ファイルの読み込み、なければつくるのみ
	if _, err := os.Stat(dir); err != nil {
		os.Mkdir(dir, 0777)
		initFile()
	}

	confb, err := ioutil.ReadFile(filename)
	if err != nil {
		initFile()
		confb, _ = ioutil.ReadFile(filename)
	}

	m := new(interface{})

	err = json.Unmarshal(confb, &m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	t.config = nestmap.New()
	t.config.Set(*m)
	t.rootTags = t.config.Child("root", "tags")
}

func (t *Tager) saveConfig() error {
	b, err := t.config.BytesIndent()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(configFile, b, 0766)
}

func (t *Tager) getTag(tag string) *nestmap.Nestmap {
	// カレントタグ用の前置処理
	if tag == "." {
		current := t.config.Child("root", "current")
		if !current.Exists() {
			fmt.Println(". を利用しましたが、カレントタグが未登録です\ntager ch -h を参照してください")
			os.Exit(1)
		}
		tag = current.ToString()
	}
	// タグ呼び出し
	cur := t.rootTags.Child(tag)
	if !cur.Exists() {
		fmt.Println(tag, "そのようなタグは存在しません")
		os.Exit(1)
	}
	return cur
}

func (t *Tager) tagAddFile(tag, glob string) {
	cur := t.getTag(tag)
	files, err := filepath.Glob(glob)
	if len(files) == 0 || err != nil {
		fmt.Println(glob, "そのようなファイルは存在しません")
		return
	}
	for _, file := range files {
		full, _ := filepath.Abs(file)
		if cur.Child("files").HasChild(full) {
			fmt.Println(file, "というファイルは既に", tag, "に登録されています")
			continue
		}
		cur.Child("files", full).Set(file)
	}
}
