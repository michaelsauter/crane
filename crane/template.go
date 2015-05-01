package crane

import (
	"fmt"
	"github.com/michaelsauter/crane/print"
	"os"
	"path"
	"regexp"
	"strings"
	"text/template"
)

type TemplateInfo struct {
	TargetedContainers []ContainerInfo
	DependencyMap      DependencyGraph
	Groups             map[string][]string
}
type TemplateInfoForContainer struct {
	CurrentContainer   ContainerInfo
	TargetedContainers []ContainerInfo
	DependencyMap      DependencyGraph
	Groups             map[string][]string
}

// Sub-interface of Container
type ContainerInfo interface {
	Name() string
	Dockerfile() string
	Image() string
	Id() string
	Dependencies() *Dependencies
	Exists() bool
	Running() bool
	Paused() bool
	ImageExists() bool
	Status() []string
}

// set of additional functions for the templates
var funcMap = template.FuncMap{
	"Contains":    strings.Contains,
	"ContainsAny": strings.ContainsAny,
	"HasPrefix":   strings.HasPrefix,
	"Join":        strings.Join,
	"Split":       strings.Split,
	"ToLower":     strings.ToLower,
	"ToTitle":     strings.ToTitle,
	"ToUpper":     strings.ToUpper,
	"MatchString": regexp.MatchString,
}

// returns a TemplateInfo from a Config object
func ConfigToTemplateInfo(c Config) TemplateInfo {
	containers := []ContainerInfo{}
	for _, container := range c.TargetedContainers() {
		containers = append(containers, container)
	}
	return TemplateInfo{containers, c.DependencyGraph(), c.Groups()}
}

// returns a TemplateInfoForContainer from a ContainerInfo and a container
func (td TemplateInfo) forContainer(currentContainer ContainerInfo) TemplateInfoForContainer {
	return TemplateInfoForContainer{currentContainer, td.TargetedContainers, td.DependencyMap, td.Groups}
}

// generates a textual output for the selected template and templateInfo
func (templateInfo TemplateInfo) template(templatePath string, outputPath string) {
	if templatePath == "" {
		print.Errorf("You must select the input template with the parameter --template (-t)\n")
		return
	}
	tmpl, err := template.New(path.Base(templatePath)).Funcs(funcMap).ParseFiles(templatePath)
	if err != nil {
		print.Errorf("Unable to parse template: %s\n", err)
		return
	}
	if strings.Contains(outputPath, "%s") {
		for _, container := range templateInfo.TargetedContainers {
			containerOutputPath := fmt.Sprintf(outputPath, container.Name())
			err = generateFile(tmpl, containerOutputPath, templateInfo.forContainer(container))
			if err != nil {
				return
			}
		}
	} else {
		generateFile(tmpl, outputPath, templateInfo)
	}
}

// generates a file with the output. If outputPath is empty, it will be dumped to os.Stdout
func generateFile(tmpl *template.Template, outputPath string, templateInfo interface{}) error {
	var err error
	writer := os.Stdout
	if outputPath != "" {
		writer, err = os.Create(outputPath)
		if err != nil {
			print.Errorf("Unable to create output file %s: %s\n", outputPath, err)
			return err
		}
		defer func() {
			writer.Close()
		}()
	}

	err = tmpl.Execute(writer, templateInfo)
	if err != nil {
		print.Errorf("ERROR: %s\n", err)
		return err
	}
	if outputPath != "" {
		print.Noticef("Generated file: %s\n", outputPath)
	}
	return nil
}

// dumps the dependency graph as a DOT to the write
func (templateInfo TemplateInfo) DOT(outputPath string) {
	const dotTemplate = `{{ $targetedContainers := .TargetedContainers }}digraph {
{{ range $name, $dependencies := .DependencyMap }}{{ with $dependencies }}  "{{ $name }}" [style=bold{{ range $targetedContainers }}{{ if eq $name .Name }},color=red{{ end }}{{ end }}]
{{ range .Link }}  "{{ $name }}"->"{{ . }}"
{{ end }}{{ range .VolumesFrom }}  "{{ $name }}"->"{{ . }}" [style=dashed]
{{ end }}{{ if ne .Net "" }}  "{{ $name }}"->"{{ .Net }}" [style=dotted]
{{ end }}{{ end }}{{ end }}}
`
	tmpl, err := template.New("dot").Parse(dotTemplate)
	if err != nil {
		print.Errorf("ERROR: %s\n", err)
		return
	}
	generateFile(tmpl, outputPath, templateInfo)
}
