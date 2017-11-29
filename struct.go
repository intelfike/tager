package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/intelfike/nestmap"
)

type Tager struct {
	// configFile string
	config   *nestmap.Nestmap
	rootTags *nestmap.Nestmap
}

// ========== init ==========

func (t *Tager) init() {
	dir, _ := filepath.Split(configFile)
	os.Mkdir(dir, 0777)
	os.Create(configFile)
	t.rootTags.MakeMap()
	t.saveConfig()
}
func (t *Tager) isInited() bool {
	return fileExists(configFile)
}

// ========== config ==========
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
	t.config.Indent = "\t"
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

// ========== get ==========
func (t *Tager) tagExists(tag string) bool {
	_, err := getTag(tag)
	return err == nil
}
func (t *Tager) getTag(tag string) (*nestmap.Nestmap, error) {
	// カレントタグ用の前置処理
	if tag == "." {
		current := t.config.Child("root", "current")
		if !current.Exists() {
			return nil, errors.New(". を利用しましたが、カレントタグが未登録です\ntager ch -h を参照してください")
		}
		tag = current.ToString()
	}
	// タグ呼び出し
	cur := t.rootTags.Child(tag)
	if !cur.Exists() {
		return nil, errors.New(tag + " そのようなタグは存在しません")
	}
	return cur, nil
}

func (t *Tager) getChildTags(tag string) ([]string, error) {
	cur, err := t.getTag(tag)
	if err != nil {
		return nil, errors.New(tag + "そのようなタグは存在しません")
	}
	ss := make([]string, 0)
	if *showFlagR {
		recNestTag(cur, tag, func(nm *nestmap.Nestmap, path string) {
			ss = append(ss, path+"/"+nm.BottomPath().(string))
		})
	} else {
		if cur.HasChild("tags") {
			ss = cur.Child("tags").Keys()
		}
	}
	return ss, nil
}
func (t *Tager) getFiles(tag string) ([]string, error) {
	cur, err := t.getTag(tag)
	if err != nil {
		return nil, errors.New(tag + "そのようなタグは存在しません")
	}
	files := make([]string, 0)
	if cur.IsMap() {
		if cur.HasChild("files") {
			files = append(files, cur.Child("files").Keys()...)
		}
		if *showFlagR {
			recNestTag(cur, "", func(nm *nestmap.Nestmap, path string) {
				if !nm.HasChild("files") {
					return
				}
				files = append(files, nm.Child("files").Keys()...)
			})
		}
	}
	return files, nil
}

// 複数のタグを指定した場合、AND計算をする
func (t *Tager) getFilesAND(tags ...string) ([]string, error) {
	files := make([][]string, 0)
	for _, v := range tags {
		fs, err := t.getFiles(v)
		if err != nil {
			return nil, errors.New(v + "そのようなタグは存在しません")
		}
		files = append(files, fs)
	}
	if len(files) == 0 {
		return nil, errors.New("該当するファイルがありませんでした")
	}
	and := make([]string, len(files[0]))
	copy(and, files[0])
	if len(files) != 1 {
		for _, v := range files[1:] {
			and = andStrings(and, v)
		}
	}
	and = uniqueStrings(and...)
	return and, nil
}

// ========== add ==========

// ファイルを追加
func (t *Tager) tagAddFile(tag string, globs ...string) {
	cur, err := t.getTag(tag)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, glob := range globs {
		files, _ := filepath.Glob(glob)
		for _, file := range files {
			full, _ := filepath.Abs(file)
			if cur.Child("files").HasChild(full) {
				fmt.Println(file, "というファイルは既に", tag, "に登録されています")
				continue
			}
			cur.Child("files", full).Set(file)
		}
	}
}

// 再帰的にファイルを追加
func (t *Tager) tagAddFileRec(tag string, globs ...string) {
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			for n, glob := range globs {
				globs[n] = path + "/" + glob
			}
			t.tagAddFile(tag, globs...)
			return nil
		}
		return nil
	})

}

// ========== autoremove ==========

func (t *Tager) autoremovableTags(tag string) ([]string, error) {
	resultTags := make([]string, 0)
	if tags, err := tager.getChildTags(tag); err == nil {
		for _, v := range tags {
			if _, err := tager.getTag(v); err == nil {
				continue
			}
			resultTags = append(resultTags, v)
		}
	} else {
		return nil, errors.New(tag + "そのようなタグは存在しません")
	}
	return resultTags, nil
}
func (t *Tager) autoremovableFiles(tag string) ([]string, error) {
	resultFiles := make([]string, 0)
	if files, err := tager.getFiles(tag); err == nil {
		for _, v := range files {
			if !fileExists(v) {
				continue
			}
			resultFiles = append(resultFiles, v)
		}
	} else {
		return nil, errors.New(tag + "そのようなタグは存在しません")
	}
	return resultFiles, nil
}
