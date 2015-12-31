package diskbuffer

import (
	"os"
	"bufio"
	"errors"
	"strconv"
	"log"
	"fmt"
	"time"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}
func filename(n int) string {
	return fileprefix + strconv.Itoa(n)
}
//
//func makeFileName() string {
//	return fileprefix + iota(100)
//}

const (
	maxdiskspce		int64 	= 1 << 30
	filesize 		int64 	= 1 << 20
	msgsize 		int64 	= 1 << 8
	fileprefix  			= "disk-buffer-tmp"
)

type diskbuffer struct  {
	inChannel	 		chan [] byte
	outChannel      	chan [] byte
	outbuffer			[][] byte
	inbuffer			[][] byte
    firstFile			int
//	firstPos			int64
	lastFile			int
//	lastPos				int64
	info 				string
}

func NewBuffer(in, out chan []byte, isNew bool) *diskbuffer {
	db := diskbuffer{
		inChannel : in,
		outChannel: out,
		inbuffer:   make([][]byte, 0, filesize),
		outbuffer:  make([][]byte, 0, filesize),
		info:       "info",
	}
	if isNew {
		infofile, err := os.Create(db.info)
		check(err)
		infofile.Close()
	}

//	db.info,_ = &os.Create("info")
//	defer db.info.Close()
	return &db
}

func (db *diskbuffer) GetInChannel() chan []byte {
	return db.inChannel
}

func (db *diskbuffer) GetOutChannel() chan []byte {
	return db.outChannel
}

func createFile(name string){
	f,err := os.Create(name)
	check(err)
	f.Close()
	fmt.Print("ha\n")
}

func (db *diskbuffer) storeTofile() (num int){
	name := filename(db.lastFile)
//	log.Println("haha"+name)
	createFile(name)
	f, err := os.OpenFile(name, os.O_RDWR|os.O_APPEND|os.O_CREATE,0644)
	log.Println(f.Stat())
	check(err)
	defer f.Close()
	count := 0
	for _,bs := range db.inbuffer {
		n,err := f.Write(bs)
		check(err)
		count += n
		f.WriteString("\n")
	}
	db.lastFile++
	return count
}

func (db *diskbuffer) Receive() {
	for {
		if db.lastFile - db.firstFile > int(maxdiskspce/(msgsize*filesize)) {
			<- db.inChannel
		} else {
			select {
			case in := <- db.inChannel:
				switch int64(len(db.inbuffer)) < filesize {
				case true:
					db.inbuffer = db.inbuffer[:len(db.inbuffer) + 1]
					db.inbuffer[len(db.inbuffer) - 1] = in
				case false:
					db.storeTofile()
					db.inbuffer = db.inbuffer[:1]
					db.inbuffer[0] = in
				}
			}
		}

	}
}

func (db *diskbuffer) Write(p []byte)  {
	go func() {
		db.inChannel <- p
	}()
}

func (db *diskbuffer) readFromSlice() error{
//	log.Println(len(db.inbuffer))
	if len(db.inbuffer) > 0 {
		db.outChannel <- db.inbuffer[0]
		db.inbuffer = db.inbuffer[1:]
		return nil
	} else {
		return errors.New("Empty Buffer")
	}

}

func (db *diskbuffer) readFirstFile() {
	name := filename(db.firstFile)
	fo, err := os.OpenFile(name, os.O_RDONLY,0600)
	check(err)
	reader := bufio.NewReader(fo)
	for {
		out, err := reader.ReadBytes('\n')
		if err == nil {
			out = out[:len(out)-1]
			db.outChannel <- out
//			log.Println(out)
		} else {
//			log.Println(err)
			break
		}
	}
	fo.Close()
	os.Remove(name)
	db.firstFile++
}

func (db *diskbuffer) readData() {
	if db.firstFile < db.lastFile {
		log.Println(db.firstFile,db.lastFile)
		db.readFirstFile()
	} else if db.firstFile == db.lastFile {
		db.readFromSlice()
	} else {
		panic(errors.New("last<first!"))
	}
}

func (db *diskbuffer) Send(){
	for  {
		db.readData()
	}
}

func (db *diskbuffer) updateinfo(f *os.File)  {
	f.Seek(0,0)
	f.WriteString(strconv.Itoa(db.firstFile) + "\n" + strconv.Itoa(db.lastFile) + "\n")
}

func (db *diskbuffer) update() {
	f, err := os.OpenFile("info", os.O_RDWR,0644)
	check(err)
	for {
		db.updateinfo(f)
		<- time.Tick(5*time.Second)
	}
}

func ifReload()  (bool, int, int) {
	fo, err := os.OpenFile("info", os.O_RDONLY,0600)
	defer fo.Close()
	if err != nil {
		return false, 0, 0
	}
	num := [2]int {0,0}
	reader := bufio.NewReader(fo)
	for i:=0; i<2; i++ {
		x, err := reader.ReadString('\n')
		if err == nil {
			x = x[:1]
			num[i],err = strconv.Atoi(x)
			check(err)
		} else {
			break
		}
	}
	log.Println(num)
	if num[0] < num[1] || (num[0] > 0 && num[1] < 0) {
		return true, num[0], num[1]
	}
	return false, 0, 0

}

func (db *diskbuffer) Reload(first,last int)  {
	log.Println("startreload")
	db.firstFile = first
	db.lastFile = last
}

func Setup(in, out chan []byte) *diskbuffer {
	var db *diskbuffer
	if x,first,last := ifReload();x {
		db = NewBuffer(in,out,false)
		db.Reload(first,last)
	} else {
		db = NewBuffer(in,out,true)
	}
	go db.Receive()
	go db.Send()
	go db.update()
	return db
}

