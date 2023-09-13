package cmd

import (
	"flag"
	"fmt"
	"github.com/Lexterl33t/omniscian-obfuscator-cfg/obfuscator"
	"os"
)

func Run() {
	var file string
	var type_obfu string

	flag.StringVar(&file, "file", "main.go", "Binary to obfuscate")
	flag.StringVar(&type_obfu, "type", "cfg", "obfuscation type")
	flag.Parse()

	if file != "" && type_obfu != "" {

		switch type_obfu {
		case "cfg":
			if err := obfuscator.CFG(file); err != nil {
				panic(err)
			}
		default:
			fmt.Println("Unknow obfuscation type")
			os.Exit(-1)
		}

	}

}
