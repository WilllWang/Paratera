package diskbuffer
import (
	"path/filepath"
	"os"
	"testing"
	"reflect"
	"fmt"
	"bufio"
	"time"
)

var bufFile = filepath.Join(os.TempDir(),"buffer-test")

func init() {
	os.Remove(bufFile)
}

func TestToSlice(t *testing.T)  {
	data := []byte("much data. such big. wow.")

	in:= make(chan []byte)
	out:= make(chan []byte)
	db := NewBuffer(in, out, true)

    go db.Receive()
	go db.Send()

	in <- data

	if !reflect.DeepEqual(db.inbuffer[0], data) {
		t.Errorf("Input data does not match")
	}

	for j:=0; j<100; j++ {
		in <- data
	}
//	log.Println(len(db.inbuffer))
	for j:=0; j<100; j++ {
		x := <- out
//		log.Println(x)
		if !reflect.DeepEqual(x, data) {
			t.Errorf("Input data does not match")
		}
	}
}

func TestToFile(t *testing.T)  {
	data := []byte("much data. such big. wow.")

	in:= make(chan []byte)
	out:= make(chan []byte)
	db := NewBuffer(in, out, true)
	db.inbuffer = db.inbuffer[:100]
	for j:=0; j<100; j++ {
		db.inbuffer[j] = data
	}
	db.storeTofile()
	ff, err := os.Open("disk-buffer-tmp0")
	defer ff.Close()
	check(err)
	reader := bufio.NewReader(ff)
	var p []byte
	p, err = reader.ReadBytes('\n')
	check(err)
	fmt.Print(p)
}

func TestReceive(t *testing.T)  {
	data := []byte("much data. such big. wow!much data. such big. wow!much data. such big. wow!much data. such big. wow!much data. such big. wow!much data. such big. wow!much data. such big. wow!much data. such big. wow!much data. such big. wow!")

	in:= make(chan []byte)
	out:= make(chan []byte)
	db := NewBuffer(in, out, true)

	go db.Receive()
	go db.Send()

	in <- data

	if !reflect.DeepEqual(db.inbuffer[0], data) {
		t.Errorf("Input data does not match")
	}

	for j:=0; j<120000; j++ {
		in <- data
	}
//	log.Println(len(db.inbuffer))
	for j:=0; j<120000; j++ {
		x := <- out
//		log.Println(x)
		if !reflect.DeepEqual(x, data) {
			t.Errorf("Input data does not match")
		}
	}
}

func TestReadInfo(t *testing.T) {
	data := []byte("much data. such big. wow!much data. such big. wow!much data. such big. wow!much data. such big. wow!much data. such big. wow!much data. such big. wow!much data. such big. wow!much data. such big. wow!much data. such big. wow!")

	in:= make(chan []byte)
	out:= make(chan []byte)
	db := NewBuffer(in, out, true)

	go db.Receive()
	go db.Send()
	go db.update()
	for j:=0; j<1200000; j++ {
		in <- data
	}
	for j:=0; j<200000; j++ {
		x := <- out
		//		log.Println(x)
		if !reflect.DeepEqual(x, data) {
			t.Errorf("Input data does not match")
		}
	}
	x,_,_ := ifReload()
	if x != true {
		t.Errorf("not return true")
	}
}

//func TestReload(t *testing.T) {
//	data := []byte("much data. such big. wow!much data. such big. wow!much data. such big. wow!much data. such big. wow!much data. such big. wow!much data. such big. wow!much data. such big. wow!much data. such big. wow!much data. such big. wow!")
//	in1:= make(chan []byte)
//	out1:= make(chan []byte)
//	Setup(in1,out1)
//	for j:=0; j<1000000; j++ {
//		x := <- out1
//		//		log.Println(x)
//		if !reflect.DeepEqual(x, data) {
//			t.Errorf("Input data does not match")
//		}
//	}
//}

func TestConcurrentInOut(t *testing.T)  {
	data := []byte("much data. such big. wow!much data. such big. wow!much data. such big. wow!much data. such big. wow!much data. such big. wow!much data. such big. wow!much data. such big. wow!much data. such big. wow!much data. such big. wow!")
	in:= make(chan []byte)
	out:= make(chan []byte)
	Setup(in,out)
	go func() {
		for j:=0; j<20000000; j++ {
			in <- data
		}
	} ()
	go func() {
		for j:=0; j<10000000; j++ {
			<- time.Tick(time.Microsecond)
			x := <- out
			//		log.Println(x)
			if !reflect.DeepEqual(x, data) {
				t.Errorf("Input data does not match")
			}
		}
	} ()
	time.Sleep(10 *time.Second)
	for j:=0; j<10000000; j++ {
		x := <- out
		//		log.Println(x)
		if !reflect.DeepEqual(x, data) {
			t.Errorf("Input data does not match")
		}
	}
}