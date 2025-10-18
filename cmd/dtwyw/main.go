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

	// reqStart, err := http.NewRequest(
	// 	"GET",
	// 	"https://api.ilovepdf.com/v1/start/compress/us",
	// 	nil,
	// )
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// reqStart.Header.Set("Authorization", token)

	// client := &http.Client{}
	// resStart, err := client.Do(reqStart)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer resStart.Body.Close()

	// var resultStart = struct {
	// 	Server           string `json:"server"`
	// 	Task             string `json:"task"`
	// 	RemainingCredits int    `json:"remaining_credits"`
	// }{}
	// if err := json.NewDecoder(resStart.Body).Decode(&resultStart); err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println(resultStart)

	// // UPLOAD

	// urlUpload := fmt.Sprintf("https://%s/v1/upload", resultStart.Server)
	// filePath := "/home/uaeh/pdf/uaeh.pdf"
	// file, err := os.Open(filePath)
	// if err != nil {
	// 	log.Fatalf("Error al abrir el archivo: %s", err)
	// }
	// defer file.Close()

	// body := &bytes.Buffer{}
	// writer := multipart.NewWriter(body)

	// part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	// if err != nil {
	// 	log.Fatalf("Error al crear la parte del formulario: %s", err)
	// }

	// _, err = io.Copy(part, file)
	// if err != nil {
	// 	log.Fatalf("Error al copiar el archivo a la parte: %s", err)
	// }

	// _ = writer.WriteField("task", resultStart.Task)

	// err = writer.Close()
	// if err != nil {
	// 	log.Fatalf("Error al cerrar el writer: %s", err)
	// }

	// reqUpload, err := http.NewRequest("POST", urlUpload, body)
	// if err != nil {
	// 	log.Fatalf("Error al crear la petición: %s", err)
	// }

	// reqUpload.Header.Set("Content-Type", writer.FormDataContentType())
	// reqUpload.Header.Set("Authorization", token)

	// respUpload, err := client.Do(reqUpload)
	// if err != nil {
	// 	log.Fatalf("Error al enviar la petición: %s", err)
	// }
	// defer respUpload.Body.Close()

	// var resultUpload = struct {
	// 	ServerFileName string `json:"server_filename"`
	// }{}
	// if err := json.NewDecoder(respUpload.Body).Decode(&resultUpload); err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println(resultUpload)

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
