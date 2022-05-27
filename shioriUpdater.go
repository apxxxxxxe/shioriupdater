package main

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const name = "shioriupdater"

var shioriPaths = [][]string{
	{"yaya.dll", "https://github.com/ponapalt/yaya-shiori/releases/latest/download/yaya.zip"},
	{"satori.dll", "https://github.com/ponapalt/satoriya-shiori/releases/latest/download/satori.zip"},
}

// {{{ downloadFile(tempDir, url string) (string, error)
func downloadFile(tempDir, url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	out, err := os.CreateTemp(tempDir, name)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return out.Name(), err
}

// }}}

// {{{ Unzip(src, dest string) error
func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		path := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			f, err := os.OpenFile(
				path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// }}}

// {{{ walkDir(dir string) ([]string, error)
func walkDir(dir string) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	var paths []string
	for _, file := range files {
		if file.IsDir() {
			paths = append(paths, walkDir(filepath.Join(dir, file.Name()))...)
			continue
		}
		paths = append(paths, filepath.Join(dir, file.Name()))
	}

	return paths
}

// }}}

// {{{ getShioriFiles() (map[string]string, error)
// shioriのdllファイルをDLしてパスを返す
func getShioriFiles(tempDir string) (map[string]string, error) {
	result := map[string]string{}

	for _, shiori := range shioriPaths {

		fileName := shiori[0]
		url := shiori[1]

		fmt.Println("ダウンロード中: ", url)

		dlPath, err := downloadFile(tempDir, url)
		if err != nil {
			return map[string]string{}, err
		}

		out := filepath.Join(filepath.Dir(dlPath), filepath.Base(dlPath)+"_out")

		if err := Unzip(dlPath, out); err != nil {
			panic(err)
		}

		files := walkDir(out)
		if err != nil {
			panic(err)
		}

		for _, f := range files {
			if strings.HasSuffix(f, fileName) {
				result[fileName] = f
				break
			}
		}

	}
	return result, nil
}

// }}}

func main() {

	var baseDir string

	if len(os.Args) > 1 {
		baseDir = os.Args[1]
	} else {
		pwd, err := os.Getwd()
		if err != nil {
			log.Fatalln(err)
		}
		baseDir = pwd
	}

	tempDir, err := os.MkdirTemp(os.TempDir(), name)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("最新版の栞を取得中...")
	dllPaths, err := getShioriFiles(tempDir)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("取得完了")

	fmt.Println("\n更新対象の検索を開始")
	files := walkDir(baseDir)
	if err != nil {
		log.Fatalln(err)
	}

	shiori := []string{}
	for _, s := range shioriPaths {
		shiori = append(shiori, s[0])
	}

	count := 0
	for _, dll := range shiori {
		srcFile, err := os.Open(dllPaths[dll])
		if err != nil {
			panic(err)
		}

		fmt.Println(dll, "を検索中...")

		for _, file := range files {
			if strings.Contains(file, dll) {

				count++
				fmt.Println("更新: ", file)

				destFile, err := os.Create(file)
				if err != nil {
					panic(err)
				}

				if _, err := io.Copy(destFile, srcFile); err != nil {
					panic(err)
				}

				destFile.Close()
			}
		}
		srcFile.Close()
	}

	if count == 0 {
		fmt.Println("更新対象が見つかりませんでした")
	} else {
		fmt.Println("検索終了")
	}

	fmt.Println("\n終了処理中...")
	if err := os.RemoveAll(tempDir); err != nil {
		panic(err)
	}

	fmt.Print("終了: Enterキーで閉じる")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

}
