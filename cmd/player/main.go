package main

import (
	"github.com/S0vina/walkmoon/internal/audio"
	"strconv"
	"fmt"
	"log"
	"os"
)

func main() {
	// Error: wrong args
	if len(os.Args) < 2 {
		fmt.Println("Uso: walkmoon <caminho_da_pasta>")
		return
	}

	// path of the folder with the musics
	folderPath := os.Args[1]

	if os.Args[2] != "0" && os.Args[2] != "1"{
		fmt.Println("No format %s format exists", os.Args[2])
	}

	mode, errMode := strconv.Atoi(os.Args[2])
	if errMode != nil {
		log.Fatal(errMode)
	}

	ap, errNewAP := audio.New() 
	if errNewAP != nil {
		log.Fatal(errNewAP)
	}

	ap.StartInputListener()

	// songs = [archives], err = any error
	playlist, errScan := ap.ScanFolder(folderPath)	

	// if ocurre some error in scanFolder func
	if errScan != nil {
		log.Fatal(errScan)
	}
	
	// if folder with no songs, break the program
	if len(playlist) == 0 {
		fmt.Println("No songs founded.")
		return
	}

	// for loop that play n songs of the folder

	if mode == 0 {
		ap.PlaySequencial(playlist)
	
	}else {
		ap.PlayShuffle(playlist)

	}
}