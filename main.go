package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/gpio", handleGPIO)
	if err := http.ListenAndServe(":80", nil); err != nil {
		log.Fatal(err)
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	mainTemplate.Execute(w, nil)
}

func handleGPIO(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = pin.Set(string(body) == "true")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

var pin = GetOutputPin(10)

type Pin struct {
	num string
}

func (r *Pin) Set(val bool) error {
	bval := []byte("0")
	if val {
		bval = []byte("1")
	}
	return ioutil.WriteFile("/sys/class/gpio/gpio"+r.num+"/value", bval, 0666)
}

func GetOutputPin(num int) *Pin {
	pin := &Pin{strconv.Itoa(num)}
	filename := "/sys/class/gpio/gpio" + pin.num
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		err := ioutil.WriteFile("/sys/class/gpio/export", []byte(pin.num), 0666)
		if err != nil {
			log.Println(err)
		}
	}
	err := ioutil.WriteFile("/sys/class/gpio/gpio"+pin.num+"/direction", []byte("out"), 0666)
	if err != nil {
		log.Println(err)
	}
	return pin
}

const mainTemplateString = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>pi gpio button</title>
    <style>
		.switch {
			position: relative;
			display: inline-block;
			width: 600px;
			height: 340px;
		}

		.switch input {display:none;}

		.slider {
			position: absolute;
			cursor: pointer;
			top: 0;
			left: 0;
			right: 0;
			bottom: 0;
			background-color: #ccc;
			-webkit-transition: .4s;
			transition: .4s;
			border-radius: 340px;
		}

		.slider:before {
			position: absolute;
			content: "";
			height: 260px;
			width: 260px;
			left: 40px;
			bottom: 40px;
			background-color: white;
			-webkit-transition: .4s;
			transition: .4s;
			border-radius: 50%;
		}

		input:checked + .slider {
			background-color: #2196F3;
		}

		input:focus + .slider {
			box-shadow: 0 0 1px #2196F3;
			border-radius: 340px;
		}

		input:checked + .slider:before {
			-webkit-transform: translateX(260px);
			-ms-transform: translateX(260px);
			transform: translateX(260px);
		}
	</style>
</head>
<body>
<label class="switch">
	<input id="butt" type="checkbox">
	<span class="slider"></span>
</label>
<script>
const butt = document.getElementById('butt')
butt.checked = false;
let to = null;
butt.onclick = ev => {
	if (butt.checked) {
		setgpio(true);
		to = setTimeout(() => {
			setgpio(false);
			butt.checked = false;
		}, 30000);
	} else {
		setgpio(false);
		if (to) {
			clearTimeout(to);
			to = null;
		}
	}
	console.log('butt' + butt.checked);
};

const setgpio = val => {
	fetch('/gpio', {method: 'PUT', body: val.toString()});
};
</script>
</body>
</html>`

var mainTemplate = template.Must(template.New("main").Parse(mainTemplateString))
