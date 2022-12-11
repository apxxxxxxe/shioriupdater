package main

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bi-zone/go-fileversion"
	"github.com/blang/semver"
	"github.com/mitchellh/go-ps"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
)

const name = "shioriupdater"
const version = "1.1.7"

var shioriPaths = [][]string{
	{"yaya.dll", "https://github.com/ponapalt/yaya-shiori/releases/latest/download/yaya.zip"},
	{"satori.dll", "https://github.com/ponapalt/satoriya-shiori/releases/latest/download/satori.zip"},
	{"ssu.dll", "https://github.com/ponapalt/satoriya-shiori/releases/latest/download/satori.zip"},
	{"satorite.exe", "https://github.com/ponapalt/satoriya-shiori/releases/latest/download/satori.zip"},
}

func getFileVersion(path string) (int, error) {
	info, err := fileversion.New(path)
	if err != nil {
		return -1, err
	}

	verString := info.FileVersion()
	verString = strings.ReplaceAll(verString, " ", "")
	verString = strings.ReplaceAll(verString, ",", "")

	ver, err := strconv.Atoi(verString)
	if err != nil {
		return -1, err
	}

	return ver, nil
}

func isProcExist(name string) bool {
	var result bool

	processes, err := ps.Processes()

	if err != nil {
		os.Exit(1)
	}

	result = false
	for _, p := range processes {
		if p.Executable() == name {
			result = true
		}
	}

	return result
}

func formatTime(t time.Time) string {
	return t.Format(time.RFC1123)
}

func updateSelf() bool {
	latest, found, err := selfupdate.DetectLatest("apxxxxxxe/shioriupdater")
	if err != nil {
		fmt.Println("エラー: ", err)
		return false
	}

	if !found {
		fmt.Println("バージョンが取得できませんでした")
		return false
	}

	v := semver.MustParse(version)
	if found && latest.Version.Equals(v) {
		fmt.Println("最新バージョンです")
		return false
	}

	if err := selfupdate.UpdateTo(latest.AssetURL, os.Args[0]); err != nil {
		log.Println("更新処理中にエラーが発生しました:", err)
		return false
	}
	fmt.Println("更新しました:", version, "->", latest.Version)
	return true
}

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

func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		modTime := f.Modified

		os.MkdirAll(dest, 0755)

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

			err = os.Chtimes(path, modTime, modTime)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func walkDir(dir string) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		switch err.(type) {
		case *fs.PathError:
			fmt.Println("エラー:", err)
			fmt.Println("このディレクトリ以下の探索をスキップします")
		}
		return []string{}
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

func getShioriFiles(tempDir string) (map[string]string, error) {
	result := map[string]string{}

	downloadedPath := map[string]string{}
	for _, shiori := range shioriPaths {

		fileName := shiori[0]
		url := shiori[1]

		var (
			unzipDir string
		)

		if downloadedPath[url] != "" {
			// ダウンロード済なら解凍先パスを取得
			unzipDir = downloadedPath[url]
		} else {
			fmt.Println("ダウンロード中:", url)

			dlPath, err := downloadFile(tempDir, url)
			if err != nil {
				return map[string]string{}, err
			}

			unzipDir = filepath.Join(filepath.Dir(dlPath), filepath.Base(dlPath)+"_out")

			if err := Unzip(dlPath, unzipDir); err != nil {
				return map[string]string{}, err
			}

			downloadedPath[url] = unzipDir
		}

		for _, f := range walkDir(unzipDir) {
			if strings.HasSuffix(f, fileName) {
				result[fileName] = f
				break
			}
		}

	}
	return result, nil
}

func main() {

	var baseDir string

	for _, args := range os.Args {
		if args == "--version" {
			fmt.Println(name, "version", version)
			return
		}
	}

	fmt.Println(name, "本体のアップデートを確認中…")
	if updateSelf() {
		fmt.Println("\n更新完了: １度閉じてから再実行してください")
		fmt.Print("終了: Enterキーで閉じる")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		return
	}
	fmt.Println("アップデート確認処理完了")

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
	defer os.RemoveAll(tempDir)

	fmt.Println("\n最新版の栞を取得中...")
	dllPaths, err := getShioriFiles(tempDir)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("取得完了")

	dllNames := []string{}
	dllBytes := map[string][]byte{}
	dllModTimes := map[string]time.Time{}
	for _, s := range shioriPaths {
		dllName := s[0]

		dllNames = append(dllNames, dllName)

		fs, err := os.Stat(dllPaths[dllName])
		if err != nil {
			log.Fatalln(err)
		}
		dllModTimes[dllName] = fs.ModTime()

		dllBytes[dllName], err = ioutil.ReadFile(dllPaths[dllName])
		if err != nil {
			log.Fatalln(err)
		}

	}

	fmt.Println("\n更新対象の検索を開始")

	count := 0

	for _, file := range walkDir(baseDir) {

		for _, dllName := range dllNames {

			if strings.Contains(file, dllName) {

				stat, err := os.Stat(file)
				if err != nil {
					log.Fatalln(err)
				}

				currentVersion, err := getFileVersion(file)
				if err != nil {
					log.Fatalln(err)
				}

				latestVersion, err := getFileVersion(dllPaths[dllName])
				if err != nil {
					log.Fatalln(err)
				}

				if currentVersion >= latestVersion {

					fmt.Println("最新版です。スキップ: ", file)

				} else {

					count++

					// ファイル内容をコピー パーミッションは旧ファイルと同等に
					if err := ioutil.WriteFile(file, dllBytes[dllName], stat.Mode().Perm()); err != nil {
						switch err.(type) {
						case *fs.PathError:
							if isProcExist("ssp.exe") {
								fmt.Println("エラー: ファイルの書き込みに失敗しました。SSPなどで開いている場合は閉じてから再実行してください。スキップ:", file)
							} else {
								fmt.Println("エラー: ファイルの書き込みに失敗しました。アプリケーションで開いている場合は閉じてから再実行してください。スキップ:", file)
							}
							continue
						default:
							log.Fatalln(err)
						}
					}

					// アクセス日時,更新日時をコピー
					if err := os.Chtimes(file, dllModTimes[dllName], dllModTimes[dllName]); err != nil {
						log.Fatalln(err)
					}

					fmt.Println("更新: ", file)

				}
			}
		}
	}

	if count == 0 {
		fmt.Println("更新対象が見つかりませんでした")
	} else {
		fmt.Println("検索終了")
	}

	fmt.Print("終了: Enterキーで閉じる")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

}
