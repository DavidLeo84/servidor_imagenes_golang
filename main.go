package main

import (
	"encoding/base64"
	"flag"
	"html/template"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type ImageData struct {
	Images   []string
	Hostname string
}

func main() {
	port := flag.String("port", "8080", "Número de puerto para el servidor web")
	imgDir := flag.String("imgDir", "images", "Directorio que contiene las imágenes")
	flag.Parse()

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "desconocido"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		images, err := getRandomImages(*imgDir, 4)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl := template.Must(template.New("index").Parse(`
        <!DOCTYPE html>
        <html lang="es">
        <head>
            <meta charset="UTF-8">
            <meta name="viewport" content="width=device-width, initial-scale=1.0">
            <title>Galería de Imágenes</title>
            <link href="https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/css/bootstrap.min.css" rel="stylesheet">
            <style>
                body {
                    background: #4AC29A;  /* fallback for old browsers */
                    background: -webkit-linear-gradient(to right, #BDFFF3, #4AC29A);  /* Chrome 10-25, Safari 5.1-6 */
                    background: rgb(2,0,36);
                    background: linear-gradient(90deg, rgba(2,0,36,1) 0%, rgba(96,9,121,1) 18%, rgba(0,212,255,1) 100%);
                }
                .modal {
                    display: none;
                    position: fixed;
                    z-index: 1;
                    left: 0;
                    top: 0;
                    width: 100%;
                    height: 100%;
                    overflow: auto;
                    background-color: rgb(0,0,0);
                    background-color: rgba(0,0,0,0.4);
                }
                .modal-content {
                    background-color: #fefefe;
                    margin: 15% auto;
                    padding: 20px;
                    border: 1px solid #888;
                    width: 80%;
                    max-width: 500px;
                }
                .close {
                    color: #aaa;
                    float: right;
                    font-size: 28px;
                    font-weight: bold;
                }
                .close:hover,
                .close:focus {
                    color: black;
                    text-decoration: none;
                    cursor: pointer;
                }
            </style>
        </head>
        <body>
            <div class="container text-center">
                <h1 class="my-4">Galería de Imágenes</h1>
                <div class="row">
                    {{range .Images}}
                    <div class="col-lg-6 col-md-6 mb-4">
                        <img class="img-fluid" style="width: 400px; height: 400px;" src="data:image;base64,{{.}}" alt="Imagen" onclick="openModal(this.src)">
                    </div>
                    {{end}}
                </div>
                <p>Servidor host: {{.Hostname}}</p>
            </div>

            <div id="myModal" class="modal">
                <div class="modal-content">
                    <span class="close" onclick="closeModal()">×</span>
                    <img id="modalImage" style="width: 500px; height: 500px;" src="" alt="Imagen">
                </div>
            </div>

            <script>
                function openModal(src) {
                    document.getElementById("modalImage").src = src;
                    document.getElementById("myModal").style.display = "block";
                }

                function closeModal() {
                    document.getElementById("myModal").style.display = "none";
                }
            </script>
        </body>
        </html>
        `))

		tmpl.Execute(w, ImageData{Images: images, Hostname: hostname})
	})

	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir(*imgDir))))

	http.ListenAndServe(":"+*port, nil)
}

func getRandomImages(dir string, count int) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var images []string
	for _, file := range files {
		if !file.IsDir() {
			ext := filepath.Ext(file.Name())
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
				imgPath := filepath.Join(dir, file.Name())
				imgBase64, err := encodeImageToBase64(imgPath)
				if err != nil {
					return nil, err
				}
				images = append(images, imgBase64)
			}
		}
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(images), func(i, j int) { images[i], images[j] = images[j], images[i] })

	if len(images) < count {
		return images, nil
	}
	return images[:count], nil
}

func encodeImageToBase64(path string) (string, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(file), nil
}
