package sender_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processloader/service/sender"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HTTPProcessSender", func() {
	var (
		testServer *httptest.Server
		baseURL    string
		s          *sender.HTTPProcessSender
		received   []byte
	)

	var sampleDef = model.ProcessDefinition{
		Name: "sample",
		Params: []model.Param{
			{Name: "p1", Mandatory: true},
		},
		Tasks: []model.Task{
			{Name: "t1", Class: "A"},
		},
	}

	When("a process is send", func() {
		BeforeEach(func() {
			testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var err error
				received, err = io.ReadAll(r.Body)
				Expect(err).To(Not(HaveOccurred()))
				Expect(r.Method).To(Equal(http.MethodPost))
				Expect(r.URL.Path).To(Equal("/startProcess"))
				w.WriteHeader(http.StatusOK)
				Expect(r.Body.Close()).To(Not(HaveOccurred()))
			}))
			baseURL = testServer.URL
			s = sender.NewHTTPProcessSender(baseURL)
		})

		AfterEach(func() {
			testServer.Close()
		})

		It("sends process successfully", func() {
			err := s.Send(sampleDef)
			Expect(err).To(BeNil())

			var parsed model.ProcessDefinition
			err = json.Unmarshal(received, &parsed)
			Expect(err).To(BeNil())
			Expect(parsed.Name).To(Equal("sample"))
			Expect(parsed.Tasks[0].Class).To(Equal("A"))
		})
	})

	When("server returns error", func() {
		BeforeEach(func() {
			testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "internal error", http.StatusInternalServerError)
			}))
			s = sender.NewHTTPProcessSender(testServer.URL)
		})

		AfterEach(func() {
			testServer.Close()
		})

		It("returns an error if status code >= 300", func() {
			err := s.Send(sampleDef)
			Expect(err).To(MatchError(ContainSubstring("API returned status 500")))
		})
	})
})
