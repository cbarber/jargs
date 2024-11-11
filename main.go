package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/goccy/go-json"
)

var jsonMap = flag.String("map", `{ "out": {{.Out|tojson}}, "in": {{.In|tojson}} }`, "JSON to map result")

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	functionMap := template.FuncMap{
		"tojson": func(data interface{}) string {
			out, err := json.Marshal(data)
			if err != nil {
				log.Fatal("Failed to encode data with `tojson`\n")
				log.Fatalf("%+v\n", data)
				os.Exit(1)
			}
			return string(out)
		},
	}

	templateArgs := make([]*template.Template, len(args))
	for i, arg := range args {
		t, err := template.New(fmt.Sprintf("arg%d", i)).Funcs(functionMap).Parse(arg)
		if err != nil {
			log.Fatalf("Failed to parse cmd as template %v: %v", arg, err)
			os.Exit(1)
		}
		templateArgs[i] = t
	}

	templateMap, err := template.New("jsonMap").Funcs(functionMap).Parse(*jsonMap)
	if err != nil {
		log.Fatalf("Failed to parse jsonMap as template %v: %v", jsonMap, err)
		os.Exit(1)
	}

	inData, err := DecodeUnknownJson(os.Stdin)
	if err != nil {
		log.Fatalf("Failed to decode stdin as json: %v", err)
		os.Exit(1)
	}

	switch inData := inData.(type) {
	case JsonArray:
		var collect []interface{}
		for _, in := range inData.Inner {
			stdout, err := ExecuteForObject(templateArgs, in)
			if err != nil {
				log.Fatalf("Failed to run command for %v: %v", in, err)
				os.Exit(1)
			}
			out := DecodeCommandStdout(stdout)
			item := MapCommandStdout(*templateMap, in, out)

			collect = append(collect, item)
		}
		output, err := json.Marshal(collect)
		if err != nil {
			fmt.Printf("Failed to encode mapped stdout as json: %v", err)
			os.Exit(1)
		}
		fmt.Print(string(output))
	case JsonObject:
		in := inData.Inner
		stdout, err := ExecuteForObject(templateArgs, in)
		if err != nil {
			log.Fatalf("Failed to run command for %v: %v", in, err)
			os.Exit(1)
		}
		out := DecodeCommandStdout(stdout)
		item := MapCommandStdout(*templateMap, in, out)

		output, err := json.Marshal(item)
		if err != nil {
			fmt.Printf("Failed to encode mapped stdout as json: %v", err)
			os.Exit(1)
		}
		fmt.Print(string(output))
	}
}

func MapCommandStdout(templateMap template.Template, in interface{}, out interface{}) interface{} {
	var buf bytes.Buffer
	data := struct {
		In  any
		Out any
	}{in, out}
	templateMap.Execute(&buf, data)

	var item interface{}
	if err := json.Unmarshal(buf.Bytes(), &item); err != nil {
		log.Printf("Could not decode jsonMap as json: %v", err)
		item = buf.String()
	}
	return item
}

func DecodeCommandStdout(stdout strings.Builder) interface{} {
	var out interface{}
	if err := json.Unmarshal([]byte(stdout.String()), &out); err != nil {
		log.Printf("Could not decode command stdout as json: %v", err)
		out = stdout.String()
	}
	return out
}

func ExecuteForObject(templateArgs []*template.Template, in interface{}) (strings.Builder, error) {
	cmd, err := BuildCommand(templateArgs, in)
	if err != nil {
		return strings.Builder{}, fmt.Errorf("BuildCommand failed with: %v", err)
	}

	var buffer strings.Builder
	cmd.Stdout = &buffer
	err = cmd.Run()
	if err != nil {
		return strings.Builder{}, fmt.Errorf("cmd.Run failed with: %v", err)
	}

	return buffer, nil
}

func BuildCommand(templateArgs []*template.Template, in any) (*exec.Cmd, error) {
	xargs := make([]string, len(templateArgs))
	data := struct{ In any }{in}
	for i, t := range templateArgs {
		var buf bytes.Buffer
		err := t.Execute(&buf, data)
		if err != nil {
			return nil, err
		}
		xargs[i] = buf.String()
	}
	cmd := exec.Command(xargs[0], xargs[1:]...)
	cmd.Stderr = os.Stderr
	return cmd, nil
}

type Json interface {
	isJson()
}

type JsonArray struct {
	Inner []interface{}
}

type JsonObject struct {
	Inner interface{}
}

func (_ JsonObject) isJson() {}
func (_ JsonArray) isJson()  {}

func DecodeUnknownJson(reader io.Reader) (Json, error) {
	decoder := json.NewDecoder(reader)
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}

	switch token := token.(type) {
	case json.Delim:
		switch token {
		case '[':
			var result JsonArray
			reader = io.MultiReader(strings.NewReader("["), decoder.Buffered())
			decoder = json.NewDecoder(reader)
			if err := decoder.Decode(&result.Inner); err != nil {
				return nil, fmt.Errorf("Failed parsing JsonArray: %v", err)
			}
			return result, nil
		case '{':
			var result JsonObject
			reader = io.MultiReader(strings.NewReader("{"), decoder.Buffered())
			decoder = json.NewDecoder(reader)
			if err := decoder.Decode(&result.Inner); err != nil {
				return nil, fmt.Errorf("Failed parsing JsonObject: %v", err)
			}
			return result, nil
		default:
			return nil, fmt.Errorf("Unexpected json delim: %v", token)
		}
	default:
		return nil, fmt.Errorf("Unexpected json token: %v", token)
	}
}
