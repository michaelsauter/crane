package crane

import (
	"io/ioutil"
	"os"
	"testing"
)

type testContainerInfo struct {
	name         string
	dockerfile   string
	image        string
	id           string
	dependencies *Dependencies
	exists       bool
	running      bool
	paused       bool
	imageExists  bool
	status       []string
}

func (c *testContainerInfo) Name() string {
	return c.name
}
func (c *testContainerInfo) Dockerfile() string {
	return c.dockerfile
}
func (c *testContainerInfo) Image() string {
	return c.image
}
func (c *testContainerInfo) Id() string {
	return c.id
}
func (c *testContainerInfo) Dependencies() *Dependencies {
	return c.dependencies
}
func (c *testContainerInfo) Exists() bool {
	return c.exists
}
func (c *testContainerInfo) Running() bool {
	return c.running
}
func (c *testContainerInfo) Paused() bool {
	return c.paused
}
func (c *testContainerInfo) ImageExists() bool {
	return c.imageExists
}
func (c *testContainerInfo) Status() []string {
	return c.status
}

func templateInfo() TemplateInfo {
	return TemplateInfo{[]ContainerInfo{containerA(), containerB()}, dependencyMap(), groups()}
}

func containerA() ContainerInfo {
	return &testContainerInfo{
		name:         "a",
		dockerfile:   "dockerfileA",
		image:        "imageA",
		id:           "idA",
		dependencies: &Dependencies{All: []string{"c"}},
		exists:       true,
		running:      true,
		paused:       true,
		imageExists:  true,
		status:       []string{"1", "2", "3", "4", "5"},
	}
}
func containerB() ContainerInfo {
	return &testContainerInfo{
		name:         "b",
		dockerfile:   "dockerfileB",
		image:        "imageB",
		id:           "idB",
		dependencies: &Dependencies{All: []string{"a", "c"}},
		exists:       false,
		running:      false,
		paused:       false,
		imageExists:  false,
		status:       []string{"6", "7", "8", "9", "10"},
	}
}

func dependencyMap() DependencyGraph {
	return DependencyGraph{
		"a": &Dependencies{All: []string{"c"}, Link: []string{"c"}},
		"b": &Dependencies{All: []string{"a", "c"}, Link: []string{"c"}, VolumesFrom: []string{"a"}},
		"c": &Dependencies{All: []string{"d"}, Net: "d"},
	}
}

func groups() map[string][]string {
	return map[string][]string{
		"group1": []string{"a", "c"},
		"group2": []string{"b", "d"},
		"group3": []string{"a", "b", "c", "d"},
	}
}

func checkErr(err error, t *testing.T) {
	if err != nil {
		t.Errorf("%v", err)
	}
}

func compareFiles(pathOutput string, pathExpected string, t *testing.T) {
	outputBytes, err := ioutil.ReadFile(pathOutput)
	checkErr(err, t)
	expectedBytes, err := ioutil.ReadFile(pathExpected)
	checkErr(err, t)

	output := string(outputBytes)
	expected := string(expectedBytes)

	if output != expected {
		t.Errorf("Invalid generated file. Expected `%v`, but got `%v`", expected, output)
	}
}

func TestTemplate(t *testing.T) {
	dirPath, err := ioutil.TempDir("./", "tempDir")
	checkErr(err, t)
	defer os.RemoveAll(dirPath)

	outputPath := dirPath + "/output.txt"
	templateInfo().template("test_templates/test.tmpl", outputPath)

	compareFiles(outputPath, "test_templates/test_output.txt", t)
}

func TestTemplateForContainer(t *testing.T) {
	dirPath, err := ioutil.TempDir("./", "tempDir")
	checkErr(err, t)
	defer os.RemoveAll(dirPath)

	outputFormat := dirPath + "/%s.txt"
	templateInfo().template("test_templates/test_per_container.tmpl", outputFormat)

	compareFiles(dirPath+"/a.txt", "test_templates/test_output_A.txt", t)
	compareFiles(dirPath+"/b.txt", "test_templates/test_output_B.txt", t)
}

func TestDOT(t *testing.T) {
	dirPath, err := ioutil.TempDir("./", "tempDir")
	checkErr(err, t)
	defer os.RemoveAll(dirPath)

	outputPath := dirPath + "/output.dot"
	templateInfo().DOT(outputPath)

	compareFiles(outputPath, "test_templates/test_graph.dot", t)
}
