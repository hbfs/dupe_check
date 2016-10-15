package main

import (
	"crypto/md5"
	//	"flag"
	"fmt"
	//"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	//"time"
)

// md5, filename, file info, last_modified
type HashMap struct {
	Hashes map[[md5.Size]byte][]os.FileInfo
}

//https://xojoc.pw/justcode/golang-file-tree-traversal.html

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

func UpdateHashMap(queue chan ReduceQueue, files map[[md5.Size]byte][]FilePathInfo, wg *sync.WaitGroup) {
	for {
		q, ok := <-queue

		if !ok {
			break
		}
		files[q.Checksum] = append(files[q.Checksum], q.FilePathInfo)

		wg.Done()
	}
	return
}

/* read files in directory
if file is dir, call recursively
if not
	pass to hasher
	hasher hashes it and sends result and sum to UpdateHashMap

Run through hashmap, print hashes with more than one file by
md5			filename		last mod date

*/

func ComputeHash(path string, info os.FileInfo, queue chan ReduceQueue, wg *sync.WaitGroup) {
	//fmt.Println("Computing: ", path)
	//defer wg.Done()

	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
	}

	defer file.Close()

	/*
		// 200% CPU, ~30 MB ram,
		// 3.35s user 105.20s system 213% cpu 50.848 total
		hash := md5.New()

		if _, err := io.Copy(hash, file); err != nil {
			fmt.Println(err)
		}

		var md5sum [md5.Size]byte
		copy(md5sum[:], hash.Sum(nil)[:16])
	*/

	//Large memory usage
	data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
	}

	md5sum := md5.Sum(data)

	queue <- ReduceQueue{md5sum, FilePathInfo{info, path}}

	return
}

func CheckDir(queue chan ReduceQueue, wg *sync.WaitGroup) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			wg.Add(1)
			go ComputeHash(path, info, queue, wg)
		}
		return nil
	}
}

func main() {

	files := make(map[[md5.Size]byte][]FilePathInfo)

	q := make(chan ReduceQueue, 32)

	var wg sync.WaitGroup

	go UpdateHashMap(q, files, &wg)

	fmt.Println("dupe_check - using", runtime.NumCPU(), "threads")

	dir := os.Args[1]

	err := filepath.Walk(dir, CheckDir(q, &wg))
	if err != nil {
		fmt.Println(err)
	}

	wg.Wait()
	//:time.Sleep(1 * time.Second)
	close(q)

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
