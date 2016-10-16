package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
)

type FilePathInfo struct {
	FileInfo os.FileInfo
	Path     string
}

type ReduceQueue struct {
	Checksum [md5.Size]byte
	FilePathInfo
}

type ByModDate []FilePathInfo

func (f ByModDate) Len() int           { return len(f) }
func (f ByModDate) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f ByModDate) Less(i, j int) bool { return f[i].FileInfo.ModTime().Before(f[j].FileInfo.ModTime()) }

func ComputeHash(path string, info os.FileInfo, files map[[md5.Size]byte][]FilePathInfo) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		fmt.Println(err)
	}
	var md5sum [md5.Size]byte
	copy(md5sum[:], hash.Sum(nil)[:16])

	files[md5sum] = append(files[md5sum], FilePathInfo{info, path})
}

func CheckDir(files map[[md5.Size]byte][]FilePathInfo) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			ComputeHash(path, info, files)
		}
		return nil
	}
}

func PrintDupes(files map[[md5.Size]byte][]FilePathInfo) {
	for k, v := range files {
		if len(v) > 1 {
			fmt.Printf("\n%x\n", k)
			sort.Sort(ByModDate(v))
			for _, w := range v {
				fmt.Println(w.FileInfo.ModTime(), "..", w.Path)
			}
		}
	}
}

func main() {
	fmt.Println("dupe_check - CPU:", runtime.NumCPU(), " / GOMAXPROCS:", runtime.GOMAXPROCS(0))

	files := make(map[[md5.Size]byte][]FilePathInfo)
	dir := os.Args[1]

	err := filepath.Walk(dir, CheckDir(files))
	if err != nil {
		fmt.Println(err)
	}

	PrintDupes(files)
}
