package main

import (
	"fmt"
	"log"
	"mynet/proto/demo"
	"net/http"
	_ "net/http/pprof"

	"github.com/ganyyy/mynet"
	"github.com/ganyyy/mynet/codec"
)

type AddReq struct {
	A, B int
}

type AddRsp struct {
	C int
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	json := codec.PBProtocol()
	json.Register(1, &demo.Req{})
	json.Register(2, &demo.Rsp{})

	server, err := mynet.Listen("tcp", "0.0.0.0:0", json, 0, mynet.HandlerFunc(func(s *mynet.Session) {
		for {
			req, err := s.Receive()
			checkErr(err)

			checkErr(s.Send(&demo.Rsp{
				Str: req.(*demo.Req).GetStr(),
			}))
		}
	}))
	checkErr(err)

	go server.Serve()

	go func() {
		http.ListenAndServe("0.0.0.0:8899", nil)
	}()

	client, err := mynet.Dial("tcp", server.Listener().Addr().String(), json, 0)
	checkErr(err)

	for i := 0; i < 10; i++ {
		err := client.Send(&demo.Req{
			Str: fmt.Sprintf("%v+%v=%v", i, i, i+i),
		})
		checkErr(err)

		log.Printf("Send:%v+%v", i, i)

		rsp, err := client.Receive()
		checkErr(err)

		log.Printf("Receive %v", rsp)
	}

}
