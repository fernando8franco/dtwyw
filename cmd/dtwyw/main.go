package main

import (
	"log"
	"os"

	"github.com/fernando8franco/dtwyw/internal/config"
)

type state struct {
	cfg *config.Config
}

func main() {
	conf, err := config.Read()
	if err != nil {
		log.Fatalf("error reading config file: %v", err)
	}

	programState := state{
		cfg: &conf,
	}

	commands := commands{
		registeredCommands: make(map[string]func(*state, command) error),
	}

	commands.Register("compress", HandlerCompress)

	if len(os.Args) < 2 {
		log.Fatal("not enough arguments were provided")
	}

	cmd := command{
		Name:      os.Args[1],
		Arguments: os.Args[2:],
	}

	err = commands.Run(&programState, cmd)
	if err != nil {
		log.Fatalf("error running the command: %v", err)
	}

	// // START

	// // UPLOAD

	// // Process

	// urlProcess := fmt.Sprintf("https://%s/v1/process", resultStart.Server)
	// type ProcessRequest struct {
	// 	Task  string `json:"task"`
	// 	Tool  string `json:"tool"`
	// 	Files []struct {
	// 		ServerFilename string `json:"server_filename"`
	// 		Filename       string `json:"filename"`
	// 	} `json:"files"`
	// }

	// data := ProcessRequest{
	// 	Task: resultStart.Task,
	// 	Tool: "compress",
	// 	Files: []struct {
	// 		ServerFilename string `json:"server_filename"`
	// 		Filename       string `json:"filename"`
	// 	}{
	// 		{ServerFilename: resultUpload.ServerFileName, Filename: "uaeh.pdf"},
	// 	},
	// }

	// jsonData, err := json.Marshal(data)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// reqProcess, err := http.NewRequest("POST", urlProcess, bytes.NewBuffer(jsonData))
	// if err != nil {
	// 	log.Fatalf("Error creating request: %s", err)
	// }

	// reqProcess.Header.Set("Content-Type", "application/json")
	// reqProcess.Header.Set("Authorization", token)

	// respProcess, err := client.Do(reqProcess)
	// if err != nil {
	// 	log.Fatalf("Error al enviar la petición: %s", err)
	// }
	// defer respProcess.Body.Close()

	// var bodyData map[string]interface{}
	// if err := json.NewDecoder(respProcess.Body).Decode(&bodyData); err != nil {
	// 	log.Fatal(err)
	// }

	// bodyJson, _ := json.MarshalIndent(bodyData, "", "  ")
	// fmt.Println("JSON response:", string(bodyJson))
}
