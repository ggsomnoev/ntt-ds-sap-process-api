package reader_test

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processloader/service/reader"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConfigReader", func() {
	var (
		tmpDir      string
		readerSvc   *reader.ConfigReader
		fileContent []byte
	)

	BeforeEach(func() {
		tmpDir = GinkgoT().TempDir()
		readerSvc = reader.NewConfigReader(tmpDir)
	})

	Describe("ReadYAMLFiles", func() {
		var expectedFiles []string

		BeforeEach(func() {
			files := map[string]string{
				"valid1.yaml": "name: proc1",
				"valid2.yml":  "name: proc2",
				"ignore.txt":  "some text",
			}
			for name, data := range files {
				path := filepath.Join(tmpDir, name)
				Expect(os.WriteFile(path, []byte(data), 0644)).To(Succeed())

				if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
					expectedFiles = append(expectedFiles, path)
				}
			}
		})

		It("returns only .yaml and .yml files", func() {
			files, err := readerSvc.ReadYAMLFiles()
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(ConsistOf(expectedFiles))
		})

		When("the directory does not exist", func() {
			It("returns an error", func() {
				badReader := reader.NewConfigReader("nonexistent")
				_, err := badReader.ReadYAMLFiles()
				Expect(err).To(MatchError(ContainSubstring("failed to read dir")))
			})
		})
	})

	Describe("ParseConfigFile", func() {
		var filePath string

		When("the file contains valid YAML", func() {
			BeforeEach(func() {
				fileContent = []byte(`name: sample-process
params:
  - name: p1
    mandatory: true
    description: some param
    default: "123"
tasks:
  - name: t1
    class: C
`)
				filePath = filepath.Join(tmpDir, "valid.yaml")
				Expect(os.WriteFile(filePath, fileContent, 0644)).To(Succeed())
			})

			It("parses the file into a ProcessDefinition", func() {
				def, err := readerSvc.ParseConfigFile(filePath)
				Expect(err).NotTo(HaveOccurred())

				Expect(def.Name).To(Equal("sample-process"))
				Expect(def.Params).To(HaveLen(1))
				Expect(def.Params[0].Name).To(Equal("p1"))
				Expect(def.Tasks).To(HaveLen(1))
				Expect(def.Tasks[0].Class).To(Equal("C"))
			})
		})

		When("the file is missing", func() {
			It("returns an error", func() {
				_, err := readerSvc.ParseConfigFile(filepath.Join(tmpDir, "missing.yaml"))
				Expect(err).To(MatchError(ContainSubstring("failed to read file")))
			})
		})

		When("the YAML is invalid", func() {
			BeforeEach(func() {
				filePath = filepath.Join(tmpDir, "bad.yaml")
				Expect(os.WriteFile(filePath, []byte("not valid yaml"), 0644)).To(Succeed())
			})

			It("returns an error", func() {
				_, err := readerSvc.ParseConfigFile(filePath)
				Expect(err).To(MatchError(ContainSubstring("failed to unmarshal file")))
			})
		})
	})
})
