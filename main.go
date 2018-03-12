package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
)

func main() {
	fmt.Println("RMdefiner0.1")

	//터미널에서 인자값 받아오기
	pDefList := flag.String("list", "define_list.txt", "치환될 문자열이 정의된 리스트파일입니다.")
	pOldList := flag.String("old", "old_list.txt", "치환작업할 원본 파일 경로 리스트파일입니다.")
	pReplace := flag.String("rp", "replace/", "치환작업후 저장할 경로입니다.")
	flag.Parse()

	fmt.Println("멀티코어 사용 등록 :", runtime.GOMAXPROCS(runtime.NumCPU()), "개")

	DefList := getDefList(pDefList)
	OldList := getOldList(pOldList)

	lenOldList := len(*OldList)
	fmt.Println("치환작업될 파일 :", lenOldList, "개")
	var wait sync.WaitGroup
	wait.Add(lenOldList)

	for _, path := range *OldList {
		go writeReplace(path, *pReplace, DefList, wait.Done)
	}

	wait.Wait()
	fmt.Println("모든 작업이 완료되었습니다.")
}

func getDefList(path *string) *map[string]string {
	defList := make(map[string]string)
	bDefList, err := ioutil.ReadFile(*path)
	if err != nil {
		fmt.Println("파일 읽기 실패")
		fmt.Println("path : ", *path)
		log.Fatalln(err)
		return &defList
	}
	sDefList := strings.FieldsFunc(string(bDefList), splitLine)
	for i, line := range sDefList {
		kv := strings.SplitN(line, "=", 2)
		old := kv[0]
		new := kv[1]
		lenOld := len(old)
		lenNew := len(new)

		if lenOld < lenNew {
			fmt.Println("치환될 문자열은 원본 문자열보다 길수 없습니다.")
			fmt.Println(i, "번 째 줄", "old:", old, "lenOld", lenOld, "new:", new, "lenNew", lenNew)
			continue
		}

		//치환 길이는 맞춰줘야
		if lenOld > lenNew {
			new += strings.Repeat(" ", lenOld-lenNew)
		}

		_, ok := defList[old]
		if ok {
			fmt.Println("중복된 원본 입니다.")
			fmt.Println(i, "번 째 줄", "old:", old, "lenOld", lenOld, "new:", new, "lenNew", lenNew)
			continue
		}
		defList[old] = new
	}
	return &defList
}

func getOldList(path *string) (oldList *[]string) {
	oldList = new([]string)
	bOldList, err := ioutil.ReadFile(*path)
	if err != nil {
		fmt.Println("파일 읽기 실패")
		fmt.Println("path : ", *path)
		log.Fatalln(err)
		return
	}
	*oldList = strings.FieldsFunc(string(bOldList), splitLine)
	return
}

func splitLine(r rune) bool {
	switch r {
	case '\r', '\n':
		return true
	}
	return false
}

func writeReplace(filePath, replacePath string, defList *map[string]string, callBack func()) {
	defer callBack()
	fmt.Println("치환 시작 :", filePath)
	name := filePath
	lastPathIndex := strings.LastIndexAny(name, "\\/")
	if lastPathIndex >= 0 {
		name = name[lastPathIndex:]
	}

	bReplace, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalln(err)
	}
	for old, new := range *defList {
		bReplace = bytes.Replace(bReplace, []byte(old), []byte(new), -1) //-1 은 무제한이라는뜻
	}
	os.MkdirAll(replacePath, 0755)
	replacePath += name
	os.Create(replacePath)
	err = ioutil.WriteFile(replacePath, bReplace, 0644)
	if err != nil {
		log.Fatalln(err)
	}
}
