package main

import (
	"log"

	git "github.com/yi-nology/git-manage-service/biz/kitex_gen/git/gitservice"
	"github.com/yi-nology/git-manage-service/biz/rpc_handler"
)

func main() {
	svr := git.NewServer(new(rpc_handler.GitServiceImpl))

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
