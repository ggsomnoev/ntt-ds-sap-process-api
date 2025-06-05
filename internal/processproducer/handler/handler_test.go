package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processproducer/handler"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processproducer/handler/handlerfakes"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrTemplate      = errors.New("template error")
	ErrValidation    = errors.New("validation error")
	ErrPublishFailed = errors.New("failed to publish process")
)

var _ = Describe("Process Handler", func() {
	var (
		e         *echo.Echo
		ctx       context.Context
		recorder  *httptest.ResponseRecorder
		publisher *handlerfakes.FakePublisher
		store     *handlerfakes.FakeProcessDefinitionStore
		reader    *handlerfakes.FakeReader
		validator *handlerfakes.FakeValidator
	)

	BeforeEach(func() {
		e = echo.New()
		ctx = context.Background()
		recorder = httptest.NewRecorder()
		publisher = &handlerfakes.FakePublisher{}
		reader = &handlerfakes.FakeReader{}
		store = &handlerfakes.FakeProcessDefinitionStore{}
		validator = &handlerfakes.FakeValidator{}
		handler.RegisterHandlers(ctx, e, publisher, reader, store, validator)
	})

	Describe("POST /startProcess", func() {
		var (
			process model.ProcessDefinition
			req     *http.Request
		)

		BeforeEach(func() {
			process = model.ProcessDefinition{
				Name: "sample",
				Tasks: []model.Task{
					{Name: "print", Class: model.LocalCmd, Parameters: map[string]string{"cmd": "echo Hello"}},
				},
			}

			body, _ := json.Marshal(handler.StartProcessRequest{
				Name: "sample",
			})

			req = httptest.NewRequest(http.MethodPost, "/startProcess", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			store.GetProcessPathByNameReturns("dir/sample.yaml", nil)
			reader.ParseConfigFileReturns(process, nil)
			reader.ApplyTemplatingToTasksReturns(process.Tasks, nil)
		})

		JustBeforeEach(func() {
			e.ServeHTTP(recorder, req)
		})

		It("succeeds", func() {
			Expect(recorder.Code).To(Equal(http.StatusOK))
			Expect(store.GetProcessPathByNameCallCount()).To(Equal(1))
			Expect(reader.ParseConfigFileCallCount()).To(Equal(1))
			Expect(reader.ApplyTemplatingToTasksCallCount()).To(Equal(1))
			Expect(validator.ValidateCallCount()).To(Equal(1))
			Expect(publisher.PublishCallCount()).To(Equal(1))
			_, actualMessage := publisher.PublishArgsForCall(0)
			Expect(actualMessage.UUID).NotTo(Equal(uuid.Nil))
			Expect(actualMessage.ProcessDefinition).To(Equal(process))
		})

		When("invalid JSON is posted", func() {
			BeforeEach(func() {
				req = httptest.NewRequest(http.MethodPost, "/startProcess", bytes.NewBufferString("{invalid json"))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			})

			It("returns 400", func() {
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			})
		})

		When("store fails to return path", func() {
			BeforeEach(func() {
				store.GetProcessPathByNameReturns("", ErrNotFound)
			})

			It("returns 400 with store error", func() {
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(ContainSubstring("No such process"))
			})
		})

		When("reader fails to parse config file", func() {
			BeforeEach(func() {
				reader.ParseConfigFileReturns(model.ProcessDefinition{}, errors.New("parse error"))
			})

			It("returns 400 with parse error", func() {
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(ContainSubstring("Process definition not found"))
			})
		})

		When("templating fails", func() {
			BeforeEach(func() {
				reader.ApplyTemplatingToTasksReturns(nil, ErrTemplate)
			})

			It("returns 400 with templating error", func() {
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(ContainSubstring("Failed to apply parameters"))
			})
		})

		When("validation fails", func() {
			BeforeEach(func() {
				validator.ValidateReturns(ErrValidation)
			})

			It("returns 400 with validation error", func() {
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(ContainSubstring("validation error"))
			})
		})

		When("publishing fails", func() {
			BeforeEach(func() {
				publisher.PublishReturns(ErrPublishFailed)
			})

			It("returns 500 with publishing error", func() {
				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
				Expect(recorder.Body.String()).To(ContainSubstring("failed to publish process"))
			})
		})
	})
})
