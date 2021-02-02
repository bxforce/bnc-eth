/*
Copyright 2020 IRT SystemX

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tool

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"text/template"
)

func SolcRender(templatePath string, contractPath string, vars map[string]interface{}) {
	templateBytes, err := ioutil.ReadFile(templatePath)
	if err != nil {
		log.Fatal(err)
	}
	t, err := template.New("tmpl").Parse(string(templateBytes))
	if err != nil {
		log.Fatal(err)
	}
	var tmplBytes bytes.Buffer
	err = t.Execute(&tmplBytes, vars)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(contractPath, []byte(tmplBytes.String()), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
}

func SolcCompile(filename string, outputDir string) {
	_, err := os.Stat(outputDir)
	if os.IsNotExist(err) {
		err = os.Mkdir(outputDir, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}
	cmd := exec.Command("solc", "--optimize", "--overwrite", "--abi", "--bin", "-o", outputDir, filename)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Println(cmd.String())
	err = cmd.Run()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
}
