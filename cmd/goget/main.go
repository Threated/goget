package main

import (
	"os"

	"github.com/Threated/goget/pkg/utils"
)

func main() {
    if len(os.Args) < 2 {
        panic("Takes at least a Url")
    }
    info, err := utils.NewRepoInfoFromUrl(os.Args[len(os.Args) - 1])
    if err != nil {
        panic(err.Error())
    }
    println(info.String())
    results := make(chan utils.Result)
    go utils.Download(info, results)
    for result := range results {
        if result.Err != nil {
            println(result.Err.Error())
            if result.Context != nil {
                println(result.Context.String())
            }
        } else {
            // println("Downloaded: " + result.Context.Name)
        }
    }


}
