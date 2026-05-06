package main

import (
	git "github.com/yi-nology/git-manage-service/biz/kitex_gen/git/gitservice"
	"github.com/yi-nology/git-manage-service/biz/rpc_handler"
	"log"
)

func main() {
	svr := git.NewServer(new(rpc_handler.GitServiceImpl))

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
